package dao

import (
	"github.com/lxn/go-pgsql"
)

const (
	QueryLoadAllAccounts = "select user_id, api_key from auth.users"
)

type BaseDao struct {
	connectionString string
}

func (self *BaseDao) getConnection() (*pgsql.Conn, error) {
	connection, err := pgsql.Connect(self.connectionString, pgsql.LogNothing)
	if nil != err {
		return nil, err
	}
	return connection, err
}

func (self *BaseDao) executeQuery(query string) (int64, error) {
	connection, err := self.getConnection()
	if err != nil {
		return 0, err
	}
	defer connection.Close()
	return connection.Execute(query)
}

func (self *BaseDao) query(query string) (*pgsql.ResultSet, error) {
	connection, err := self.getConnection()
	if err != nil {
		return nil, err
	}
	return connection.Query(query)
}

func NewBaseDao(connectionString string) (*BaseDao, error) {
	var err error
	instance := new(BaseDao)
	instance.connectionString = connectionString
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (self *BaseDao) Close() error {
	return nil
}

type PgAccountsDao struct {
	*BaseDao
}

func NewPgAccountsDao(connString string) *PgAccountsDao {
	return &PgAccountsDao{&BaseDao{connString}}
}

func (self *PgAccountsDao) LoadAccountsByApiKey(dest ApiKeyMap) error {
	rs, err := self.query(QueryLoadAllAccounts)
	if err != nil {
		return err
	}
	defer rs.Close()

	for {
		hasRow, err := rs.FetchNext()
		if err != nil {
			return err
		}
		if hasRow {
			userId, _, _ := rs.Int64(0)
			api_key, _, _ := rs.String(1)
			dest[api_key] = NewAccountInfo(userId)
		} else {
			hasResult, err := rs.NextResult()
			if err != nil {
				return err
			}
			if !hasResult {
				break
			}
		}
	}
	return nil
}
