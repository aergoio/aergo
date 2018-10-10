package contract

/*
#include "sqlite3-binding.h"
*/
import "C"
import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var (
	ErrDBOpen = errors.New("failed to open the sql database")
	ErrUndo   = errors.New("failed to undo the sql database")
	ErrFindRp = errors.New("cannot find a recover point")

	database = &Database{}
	load     sync.Once

	logger = log.NewLogger("statesql")

	queryConn     *SQLiteConn
	queryConnLock sync.Mutex
)

const (
	statesqlDriver = "statesql"
	queryDriver    = "query"
)

type Database struct {
	DBs        map[types.AccountID]*DB
	OpenDbName types.AccountID
	DataDir    string
}

func init() {
	sql.Register(statesqlDriver, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if _, ok := database.DBs[database.OpenDbName]; !ok {
				database.DBs[database.OpenDbName] = &DB{
					Conn: nil,
					db:   nil,
					tx:   nil,
					conn: conn,
					name: database.OpenDbName,
				}
			} else {
				logger.Warn().Msg("duplicated connection")
			}
			return nil
		},
	})
	sql.Register(queryDriver, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			queryConn = conn
			return nil
		},
	})
}

func checkPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
	}
	return err
}

func LoadDatabase(dataDir string) error {
	var err error
	load.Do(func() {
		path := filepath.Join(dataDir, statesqlDriver)
		logger.Debug().Str("path", path).Msg("loading statesql")
		if err = checkPath(path); err == nil {
			database.DBs = make(map[types.AccountID]*DB)
			database.DataDir = path
		}
	})
	return err
}

func CloseDatabase() {
	for _, db := range database.DBs {
		_ = db.close()
	}
}

func SaveRecoveryPoint(sdb *state.ChainStateDB, bs *state.BlockState) error {
	for id, db := range database.DBs {
		if db.tx != nil {
			err := db.tx.Commit()
			db.tx = nil
			if err != nil {
				return err
			}
			rp := db.recoveryPoint()
			if rp == 0 {
				return ErrFindRp
			}
			if rp > 0 {
				logger.Debug().Str("db_name", id.String()).Uint64("commit_id", rp).Msg("save recovery point")
				receiverState, err := sdb.GetBlockAccountClone(bs, id)
				if err != nil {
					return err
				}
				receiverChange := types.Clone(*receiverState).(types.State)
				receiverChange.SqlRecoveryPoint = uint64(rp)
				bs.PutAccount(id, &receiverChange)
			}
		}
	}
	return nil
}

func BeginTx(dbName types.AccountID, rp uint64) (Tx, error) {
	db, err := conn(dbName)
	if err != nil {
		return nil, err
	}
	return db.beginTx(rp)
}

func BeginReadOnly(dbName types.AccountID, rp uint64) (Tx, error) {
	db, err := readOnlyConn(dbName)
	if err != nil {
		return nil, err
	}
	return newReadOnlyTx(db, rp)
}

func conn(dbName types.AccountID) (*DB, error) {
	if db, ok := database.DBs[dbName]; ok {
		return db, nil
	}
	return openDB(dbName)
}

func dataSrc(dbName string) string {
	return fmt.Sprintf("file:%s/%s.db?branches=on", database.DataDir, dbName)
}

func readOnlyConn(dbName types.AccountID) (*DB, error) {
	queryConnLock.Lock()
	defer queryConnLock.Unlock()

	db, err := sql.Open(queryDriver, dataSrc(dbName.String())+"&_query_only=true")
	if err != nil {
		return nil, ErrDBOpen
	}
	err = db.Ping()
	if err != nil {
		logger.Fatal().Err(err)
		_ = db.Close()
		return nil, ErrDBOpen
	}
	c, err := db.Conn(context.Background())
	if err != nil {
		logger.Fatal().Err(err)
		_ = db.Close()
		return nil, ErrDBOpen
	}
	return &DB{
		Conn: c,
		db:   db,
		tx:   nil,
		conn: queryConn,
		name: dbName,
	}, nil
}

func openDB(dbName types.AccountID) (*DB, error) {
	database.OpenDbName = dbName
	db, err := sql.Open(statesqlDriver, dataSrc(dbName.String()))
	if err != nil {
		return nil, ErrDBOpen
	}
	c, err := db.Conn(context.Background())
	if err != nil {
		logger.Fatal().Err(err)
		_ = db.Close()
		return nil, ErrDBOpen
	}
	err = c.PingContext(context.Background())
	if err != nil {
		logger.Fatal().Err(err)
		_ = c.Close()
		_ = db.Close()
		return nil, ErrDBOpen
	}
	_, err = c.ExecContext(context.Background(), "create table if not exists _dummy(_dummy)")
	if err != nil {
		logger.Fatal().Err(err)
		_ = c.Close()
		_ = db.Close()
		return nil, ErrDBOpen
	}
	database.DBs[dbName].Conn = c
	database.DBs[dbName].db = db
	return database.DBs[dbName], nil
}

type DB struct {
	*sql.Conn
	db   *sql.DB
	tx   Tx
	conn *SQLiteConn
	name types.AccountID
}

func (db *DB) beginTx(rp uint64) (Tx, error) {
	logger.Debug().Str("db_name", db.name.String()).Msg("begin transaction")
	if db.tx == nil {
		err := db.restoreRecoveryPoint(rp)
		if err != nil {
			return nil, err
		}
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			return nil, err
		}
		db.tx = &WritableTx{
			TxCommon: TxCommon{db: db},
			Tx:       tx,
		}
	}
	return db.tx, nil
}

type branchInfo struct {
	TotalCommits uint64 `json:"total_commits"`
}

func (db *DB) recoveryPoint() uint64 {
	row := db.QueryRowContext(context.Background(), "pragma branch_info(master)")
	var rv string
	err := row.Scan(&rv)
	if err != nil {
		return uint64(0)
	}
	var bi branchInfo
	err = json.Unmarshal([]byte(rv), &bi)
	if err != nil {
		return uint64(0)
	}
	return bi.TotalCommits
}

func (db *DB) restoreRecoveryPoint(stateRp uint64) error {
	lastRp := db.recoveryPoint()
	logger.Debug().Str("db_name", db.name.String()).Uint64("state_rp", stateRp).
		Uint64("last_rp", lastRp).Msgf("restore recovery point, %p", db.Conn)
	if lastRp == 0 {
		return ErrFindRp
	}
	if stateRp == lastRp {
		return nil
	}
	if stateRp > lastRp {
		return ErrUndo
	}
	if err := db.rollbackToRecoveryPoint(stateRp); err != nil {
		return err
	}
	logger.Debug().Str("db_name", db.name.String()).Uint64("commit_id", stateRp).Msg(
		"restore recovery point",
	)
	return nil
}

func (db *DB) rollbackToRecoveryPoint(rp uint64) error {
	_, err := db.ExecContext(
		context.Background(),
		fmt.Sprintf("pragma branch_truncate(master.%d)", rp),
	)
	return err
}

func (db *DB) snapshotView(rp uint64) error {
	logger.Debug().Uint64("rp", rp).Msgf("snapshot view, %p", db.Conn)
	_, err := db.ExecContext(
		context.Background(),
		fmt.Sprintf("pragma branch=master.%d", rp),
	)
	return err
}

func (db *DB) close() error {
	err := db.Conn.Close()
	if err != nil {
		_ = db.db.Close()
		return err
	}
	return db.db.Close()
}

type Tx interface {
	Commit() error
	Rollback() error
	Savepoint() error
	Release() error
	RollbackToSavepoint() error
	GetHandle() *C.sqlite3
}

type TxCommon struct {
	db *DB
}

func (tx *TxCommon) GetHandle() *C.sqlite3 {
	return tx.db.conn.db
}

type WritableTx struct {
	TxCommon
	*sql.Tx
}

func (tx *WritableTx) Commit() error {
	logger.Debug().Str("db_name", tx.db.name.String()).Msg("commit")
	return tx.Tx.Commit()
}

func (tx *WritableTx) Rollback() error {
	logger.Debug().Str("db_name", tx.db.name.String()).Msg("rollback")
	return tx.Tx.Rollback()
}

func (tx *WritableTx) Savepoint() error {
	logger.Debug().Str("db_name", tx.db.name.String()).Msg("savepoint")
	_, err := tx.Tx.Exec("SAVEPOINT \"" + tx.db.name.String() + "\"")
	return err
}

func (tx *WritableTx) Release() error {
	logger.Debug().Str("db_name", tx.db.name.String()).Msg("release")
	_, err := tx.Tx.Exec("RELEASE SAVEPOINT \"" + tx.db.name.String() + "\"")
	return err
}

func (tx *WritableTx) RollbackToSavepoint() error {
	logger.Debug().Str("db_name", tx.db.name.String()).Msg("rollback to savepoint")
	_, err := tx.Tx.Exec("ROLLBACK TO SAVEPOINT \"" + tx.db.name.String() + "\"")
	return err
}

type ReadOnlyTx struct {
	TxCommon
}

func newReadOnlyTx(db *DB, rp uint64) (Tx, error) {
	if err := db.snapshotView(rp); err != nil {
		return nil, err
	}
	tx := &ReadOnlyTx{
		TxCommon: TxCommon{db: db},
	}
	return tx, nil
}

func (tx *ReadOnlyTx) Commit() error {
	return errors.New("only select queries allowed")
}

func (tx *ReadOnlyTx) Rollback() error {
	logger.Debug().Str("db_name", tx.db.name.String()).Msg("read-only tx is closed")
	return tx.db.close()
}

func (tx *ReadOnlyTx) Savepoint() error {
	return errors.New("only select queries allowed")
}

func (tx *ReadOnlyTx) Release() error {
	return errors.New("only select queries allowed")
}

func (tx *ReadOnlyTx) RollbackToSavepoint() error {
	return errors.New("only select queries allowed")
}
