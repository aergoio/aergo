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
	"github.com/aergoio/aergo/internal/enc"
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
	DBs        map[string]*DB
	OpenDbName string
	DataDir    string
}

func init() {
	sql.Register(statesqlDriver, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if _, ok := database.DBs[database.OpenDbName]; !ok {
				b, err := enc.ToBytes(database.OpenDbName)
				if err != nil {
					logger.Error().Err(err).Msg("Open SQL Connection")
					return nil
				}
				database.DBs[database.OpenDbName] = &DB{
					Conn:      nil,
					db:        nil,
					tx:        nil,
					conn:      conn,
					name:      database.OpenDbName,
					accountID: types.AccountID(types.ToHashID(b)),
				}
			} else {
				logger.Warn().Err(errors.New("duplicated connection")).Msg("Open SQL Connection")
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
			database.DBs = make(map[string]*DB)
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

func SaveRecoveryPoint(bs *state.BlockState) error {
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
				if logger.IsDebugEnabled() {
					logger.Debug().Str("db_name", id).Uint64("commit_id", rp).Msg("save recovery point")
				}
				receiverState, err := bs.GetAccountState(db.accountID)
				if err != nil {
					return err
				}
				receiverChange := types.State(*receiverState)
				receiverChange.SqlRecoveryPoint = uint64(rp)
				err = bs.PutState(db.accountID, &receiverChange)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func BeginTx(dbName string, rp uint64) (Tx, error) {
	db, err := conn(dbName)
	if err != nil {
		return nil, err
	}
	return db.beginTx(rp)
}

func BeginReadOnly(dbName string, rp uint64) (Tx, error) {
	db, err := readOnlyConn(dbName)
	if err != nil {
		return nil, err
	}
	return newReadOnlyTx(db, rp)
}

func conn(dbName string) (*DB, error) {
	if db, ok := database.DBs[dbName]; ok {
		return db, nil
	}
	return openDB(dbName)
}

func dataSrc(dbName string) string {
	return fmt.Sprintf("file:%s/%s.db?branches=on", database.DataDir, dbName)
}

func readOnlyConn(dbName string) (*DB, error) {
	queryConnLock.Lock()
	defer queryConnLock.Unlock()

	db, err := sql.Open(queryDriver, dataSrc(dbName)+"&_query_only=true")
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

func openDB(dbName string) (*DB, error) {
	database.OpenDbName = dbName
	db, err := sql.Open(statesqlDriver, dataSrc(dbName))
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
	db        *sql.DB
	tx        Tx
	conn      *SQLiteConn
	name      string
	accountID types.AccountID
}

func (db *DB) beginTx(rp uint64) (Tx, error) {
	if db.tx == nil {
		err := db.restoreRecoveryPoint(rp)
		if err != nil {
			return nil, err
		}
		if logger.IsDebugEnabled() {
			logger.Debug().Str("db_name", db.name).Msg("begin transaction")
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
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", db.name).
			Uint64("state_rp", stateRp).
			Uint64("last_rp", lastRp).Msgf("restore recovery point")
	}
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
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", db.name).Uint64("commit_id", stateRp).
			Msg("restore recovery point")
	}
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
	if logger.IsDebugEnabled() {
		logger.Debug().Uint64("rp", rp).Msgf("snapshot view, %p", db.Conn)
	}
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
	SubSavepoint(string) error
	SubRelease(string) error
	RollbackToSubSavepoint(string) error
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
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", tx.db.name).Msg("commit")
	}
	return tx.Tx.Commit()
}

func (tx *WritableTx) Rollback() error {
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", tx.db.name).Msg("rollback")
	}
	return tx.Tx.Rollback()
}

func (tx *WritableTx) Savepoint() error {
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", tx.db.name).Msg("savepoint")
	}
	_, err := tx.Tx.Exec("SAVEPOINT \"" + tx.db.name + "\"")
	return err
}

func (tx *WritableTx) SubSavepoint(name string) error {
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", name).Msg("savepoint")
	}
	_, err := tx.Tx.Exec("SAVEPOINT \"" + name + "\"")
	return err
}

func (tx *WritableTx) Release() error {
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", tx.db.name).Msg("release")
	}
	_, err := tx.Tx.Exec("RELEASE SAVEPOINT \"" + tx.db.name + "\"")
	return err
}

func (tx *WritableTx) SubRelease(name string) error {
	if logger.IsDebugEnabled() {
		logger.Debug().Str("name", name).Msg("release")
	}
	_, err := tx.Tx.Exec("RELEASE SAVEPOINT \"" + name + "\"")
	return err
}

func (tx *WritableTx) RollbackToSavepoint() error {
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", tx.db.name).Msg("rollback to savepoint")
	}
	_, err := tx.Tx.Exec("ROLLBACK TO SAVEPOINT \"" + tx.db.name + "\"")
	return err
}

func (tx *WritableTx) RollbackToSubSavepoint(name string) error {
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", name).Msg("rollback to savepoint")
	}
	_, err := tx.Tx.Exec("ROLLBACK TO SAVEPOINT \"" + name + "\"")
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
	if logger.IsDebugEnabled() {
		logger.Debug().Str("db_name", tx.db.name).Msg("read-only tx is closed")
	}
	return tx.db.close()
}

func (tx *ReadOnlyTx) Savepoint() error {
	return errors.New("only select queries allowed")
}

func (tx *ReadOnlyTx) Release() error {
	return errors.New("only select queries allowed")
}

func (tx *ReadOnlyTx) RollbackToSavepoint() error {
	return tx.Rollback()
}

func (tx *ReadOnlyTx) SubSavepoint(name string) error {
	return nil
}

func (tx *ReadOnlyTx) SubRelease(name string) error {
	return nil
}

func (tx *ReadOnlyTx) RollbackToSubSavepoint(name string) error {
	return nil
}
