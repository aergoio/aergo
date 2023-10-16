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
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

var (
	ErrDBOpen = errors.New("failed to open the sql database")
	ErrUndo   = errors.New("failed to undo the sql database")
	ErrFindRp = errors.New("cannot find a recovery point")

	database = &sqlDatabase{}
	load     sync.Once

	sqlLgr = log.NewLogger("statesql")

	queryConn     *SQLiteConn
	queryConnLock sync.Mutex
)

const (
	statesqlDriver = "statesql"
	queryDriver    = "query"
)

type sqlDatabase struct {
	DBs        map[string]*litetree
	OpenDbName string
	DataDir    string
}

func init() {
	sql.Register(statesqlDriver, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if _, ok := database.DBs[database.OpenDbName]; !ok {
				b, err := enc.ToBytes(database.OpenDbName)
				if err != nil {
					sqlLgr.Error().Err(err).Msg("Open SQL Connection")
					return nil
				}
				database.DBs[database.OpenDbName] = &litetree{
					Conn:      nil,
					db:        nil,
					tx:        nil,
					conn:      conn,
					name:      database.OpenDbName,
					accountID: types.AccountID(types.ToHashID(b)),
				}
			} else {
				sqlLgr.Warn().Err(errors.New("duplicated connection")).Msg("Open SQL Connection")
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
		sqlLgr.Debug().Str("path", path).Msg("loading statesql")
		if err = checkPath(path); err == nil {
			database.DBs = make(map[string]*litetree)
			database.DataDir = path
		}
	})
	return err
}

func LoadTestDatabase(dataDir string) error {
	var err error
	path := filepath.Join(dataDir, statesqlDriver)
	sqlLgr.Debug().Str("path", path).Msg("loading statesql")
	if err = checkPath(path); err == nil {
		database.DBs = make(map[string]*litetree)
		database.DataDir = path
	}
	return err
}

func CloseDatabase() {
	var err error
	for name, db := range database.DBs {
		if db.tx != nil {
			err = db.tx.rollback()
			if err != nil {
				sqlLgr.Warn().Err(err).Str("db_name", name).Msg("SQL TX close")
			}
			db.tx = nil
		}
		err = db.close()
		if err != nil {
			sqlLgr.Warn().Err(err).Str("db_name", name).Msg("SQL DB close")
		}
		delete(database.DBs, name)
	}
}

func SaveRecoveryPoint(bs *state.BlockState) error {
	defer CloseDatabase()

	for id, db := range database.DBs {
		if db.tx != nil {
			err := db.tx.commit()
			db.tx = nil
			if err != nil {
				sqlLgr.Warn().Err(err).Str("db_name", id).Msg("SQL TX commit")
				continue
			}
			rp := db.recoveryPoint()
			if rp == 0 {
				return ErrFindRp
			}
			if rp > 0 {
				if sqlLgr.IsDebugEnabled() {
					sqlLgr.Debug().Str("db_name", id).Uint64("commit_id", rp).Msg("save recovery point")
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

func beginTx(dbName string, rp uint64) (sqlTx, error) {
	db, err := conn(dbName)
	defer func() {
		if err != nil {
			delete(database.DBs, dbName)
		}
	}()
	if err != nil {
		return nil, err
	}
	if rp == 1 {
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			goto failed
		}
		_, err = tx.ExecContext(context.Background(), "create table if not exists _dummy(_dummy)")
		if err != nil {
			goto failed
		}
		err = tx.Commit()
		if err != nil {
			goto failed
		}
	}
failed:
	if err != nil {
		sqlLgr.Fatal().Err(err)
		_ = db.close()
		return nil, ErrDBOpen
	}
	return db.beginTx(rp)
}

func beginReadOnly(dbName string, rp uint64) (sqlTx, error) {
	db, err := readOnlyConn(dbName)
	if err != nil {
		return nil, err
	}
	return newReadOnlySqlTx(db, rp)
}

func conn(dbName string) (*litetree, error) {
	if db, ok := database.DBs[dbName]; ok {
		return db, nil
	}
	return openDB(dbName)
}

func dataSrc(dbName string) string {
	return fmt.Sprintf(
		"file:%s/%s.db?branches=on&max_db_size=%d",
		database.DataDir,
		dbName,
		maxSQLDBSize*1024*1024)
}

func readOnlyConn(dbName string) (*litetree, error) {
	queryConnLock.Lock()
	defer queryConnLock.Unlock()

	db, err := sql.Open(queryDriver, dataSrc(dbName)+"&_query_only=true")
	if err != nil {
		sqlLgr.Fatal().Err(err)
		return nil, ErrDBOpen
	}
	var c *sql.Conn
	err = db.Ping()
	if err == nil {
		c, err = db.Conn(context.Background())
	}
	if err != nil {
		sqlLgr.Fatal().Err(err)
		_ = db.Close()
		return nil, ErrDBOpen
	}
	return &litetree{
		Conn: c,
		db:   db,
		tx:   nil,
		conn: queryConn,
		name: dbName,
	}, nil
}

func openDB(dbName string) (*litetree, error) {
	database.OpenDbName = dbName
	db, err := sql.Open(statesqlDriver, dataSrc(dbName))
	if err != nil {
		sqlLgr.Fatal().Err(err)
		return nil, ErrDBOpen
	}
	c, err := db.Conn(context.Background())
	if err != nil {
		sqlLgr.Fatal().Err(err)
		_ = db.Close()
		return nil, ErrDBOpen
	}
	err = c.PingContext(context.Background())
	if err != nil {
		sqlLgr.Fatal().Err(err)
		_ = c.Close()
		_ = db.Close()
		return nil, ErrDBOpen
	}
	database.DBs[dbName].Conn = c
	database.DBs[dbName].db = db
	return database.DBs[dbName], nil
}

type litetree struct {
	*sql.Conn
	db        *sql.DB
	tx        sqlTx
	conn      *SQLiteConn
	name      string
	accountID types.AccountID
}

func (db *litetree) beginTx(rp uint64) (sqlTx, error) {
	if db.tx == nil {
		err := db.restoreRecoveryPoint(rp)
		if err != nil {
			return nil, err
		}
		if sqlLgr.IsDebugEnabled() {
			sqlLgr.Debug().Str("db_name", db.name).Msg("begin transaction")
		}
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			return nil, err
		}
		db.tx = &writableSqlTx{
			sqlTxCommon: sqlTxCommon{litetree: db},
			Tx:          tx,
		}
	}
	return db.tx, nil
}

type branchInfo struct {
	TotalCommits uint64 `json:"total_commits"`
}

func (db *litetree) recoveryPoint() uint64 {
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

func (db *litetree) restoreRecoveryPoint(stateRp uint64) error {
	lastRp := db.recoveryPoint()
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", db.name).
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
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", db.name).Uint64("commit_id", stateRp).
			Msg("restore recovery point")
	}
	return nil
}

func (db *litetree) rollbackToRecoveryPoint(rp uint64) error {
	_, err := db.ExecContext(
		context.Background(),
		fmt.Sprintf("pragma branch_truncate(master.%d)", rp),
	)
	return err
}

func (db *litetree) snapshotView(rp uint64) error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Uint64("rp", rp).Msgf("snapshot view, %p", db.Conn)
	}
	_, err := db.ExecContext(
		context.Background(),
		fmt.Sprintf("pragma branch=master.%d", rp),
	)
	if err != nil && rp == 1 {
		return nil
	}
	return err
}

func (db *litetree) close() error {
	err := db.Conn.Close()
	if err != nil {
		_ = db.db.Close()
		return err
	}
	return db.db.Close()
}

type sqlTx interface {
	commit() error
	rollback() error
	savepoint() error
	release() error
	rollbackToSavepoint() error
	subSavepoint(string) error
	subRelease(string) error
	rollbackToSubSavepoint(string) error
	getHandle() *C.sqlite3
	close() error
	begin() error
}

type sqlTxCommon struct {
	*litetree
}

func (tx *sqlTxCommon) getHandle() *C.sqlite3 {
	return tx.litetree.conn.db
}

type writableSqlTx struct {
	sqlTxCommon
	*sql.Tx
}

var _ sqlTx = (*writableSqlTx)(nil)

func (tx *writableSqlTx) commit() error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", tx.litetree.name).Msg("commit")
	}
	return tx.Tx.Commit()
}

func (tx *writableSqlTx) rollback() error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", tx.litetree.name).Msg("rollback")
	}
	return tx.Tx.Rollback()
}

func (tx *writableSqlTx) savepoint() error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", tx.litetree.name).Msg("savepoint")
	}
	_, err := tx.Tx.Exec("SAVEPOINT \"" + tx.litetree.name + "\"")
	return err
}

func (tx *writableSqlTx) subSavepoint(name string) error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", name).Msg("savepoint")
	}
	_, err := tx.Tx.Exec("SAVEPOINT \"" + name + "\"")
	return err
}

func (tx *writableSqlTx) release() error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", tx.litetree.name).Msg("release")
	}
	err := tx.litetree.conn.DBCacheFlush()
	if err != nil {
		return err
	}
	_, err = tx.Tx.Exec("RELEASE SAVEPOINT \"" + tx.litetree.name + "\"")
	return err
}

func (tx *writableSqlTx) subRelease(name string) error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("name", name).Msg("release")
	}
	_, err := tx.Tx.Exec("RELEASE SAVEPOINT \"" + name + "\"")
	return err
}

func (tx *writableSqlTx) rollbackToSavepoint() error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", tx.litetree.name).Msg("rollback to savepoint")
	}
	_, err := tx.Tx.Exec("ROLLBACK TO SAVEPOINT \"" + tx.litetree.name + "\"")
	return err
}

func (tx *writableSqlTx) rollbackToSubSavepoint(name string) error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", name).Msg("rollback to savepoint")
	}
	_, err := tx.Tx.Exec("ROLLBACK TO SAVEPOINT \"" + name + "\"")
	return err
}

func (tx *writableSqlTx) close() error {
	return errors.New("assert(only read-tx allowed)")
}

func (tx *writableSqlTx) begin() error {
	_, err := tx.Tx.Exec("BEGIN")
	return err
}

type readOnlySqlTx struct {
	sqlTxCommon
}

var _ sqlTx = (*readOnlySqlTx)(nil)

func newReadOnlySqlTx(db *litetree, rp uint64) (sqlTx, error) {
	if err := db.snapshotView(rp); err != nil {
		return nil, err
	}
	tx := &readOnlySqlTx{
		sqlTxCommon: sqlTxCommon{litetree: db},
	}
	return tx, nil
}

func (tx *readOnlySqlTx) commit() error {
	return errors.New("only select queries allowed")
}

func (tx *readOnlySqlTx) rollback() error {
	if sqlLgr.IsDebugEnabled() {
		sqlLgr.Debug().Str("db_name", tx.litetree.name).Msg("read-only tx is closed")
	}
	return tx.litetree.close()
}

func (tx *readOnlySqlTx) savepoint() error {
	return errors.New("only select queries allowed")
}

func (tx *readOnlySqlTx) release() error {
	return errors.New("only select queries allowed")
}

func (tx *readOnlySqlTx) rollbackToSavepoint() error {
	return tx.rollback()
}

func (tx *readOnlySqlTx) subSavepoint(name string) error {
	return nil
}

func (tx *readOnlySqlTx) subRelease(name string) error {
	return nil
}

func (tx *readOnlySqlTx) rollbackToSubSavepoint(name string) error {
	return nil
}

func (tx *readOnlySqlTx) close() error {
	return tx.sqlTxCommon.close()
}

func (tx *readOnlySqlTx) begin() error {
	return errors.New("assert(only writable-tx allowed)")
}
