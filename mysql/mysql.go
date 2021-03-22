package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"time"
)

type MySQLConf struct {
	Addr  string `json:"addr" yaml:"addr"`
	Max   int    `json:"max" yaml:"max"`
	Idle  int    `json:"idle" yaml:"idle"`
	Debug bool   `json:"debug" yaml:"debug"`
}
type MySQLClient struct {
	*xorm.Engine
	Config *MySQLConf
}
type ParamAnd struct {
	Col       string      `json:"col"`
	Condition string      `json:"condition"`
	Value     interface{} `json:"value"`
}
type PageParam struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func NewMySQLClient(conf *MySQLConf) (*MySQLClient, error) {
	db, err := xorm.NewEngine("mysql", conf.Addr)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(conf.Idle)
	db.SetMaxOpenConns(conf.Max)
	db.SetConnMaxLifetime(time.Hour)
	db.ShowSQL(conf.Debug)
	db.Logger().SetLevel(1)
	return &MySQLClient{db, conf}, nil
}

func (c *MySQLClient) GetByParamAnd(parmas []*ParamAnd, obj interface{}) (bool, error) {
	switch len(parmas) {
	case 0:
		return c.Get(obj)
	case 1:
		return c.Where(fmt.Sprintf("%v %s ?", parmas[0].Col, parmas[0].Condition), parmas[0].Value).Get(obj)
	default:
		sess := c.NewSession()
		defer sess.Close()
		return sessionAnd(sess, parmas).Get(obj)
	}
}
func (c *MySQLClient) GetBySessionParamAnd(sess *xorm.Session, parmas []*ParamAnd, obj interface{}) (bool, error) {
	return sessionAnd(sess, parmas).Get(obj)
}
func (c *MySQLClient) FindByParamAnd(parmas []*ParamAnd, page *PageParam, obj interface{}) (int64, error) {
	switch len(parmas) {
	case 0:
		return c.Limit(page.Limit, page.Offset).FindAndCount(obj)
	case 1:
		return c.Where(fmt.Sprintf("%v %s ?", parmas[0].Col, parmas[0].Condition), parmas[0].Value).Limit(page.Limit, page.Offset).FindAndCount(obj)
	default:
		sess := c.NewSession()
		defer sess.Close()
		return sessionAnd(sess, parmas).Limit(page.Limit, page.Offset).FindAndCount(obj)
	}
}
func FindBySessionParamAnd(sess *xorm.Session, parmas []*ParamAnd, page *PageParam, obj interface{}) (int64, error) {
	return sessionAnd(sess, parmas).Limit(page.Limit, page.Offset).FindAndCount(obj)
}

func InsertBySession(sess *xorm.Session, obj interface{}) error {
	_, err := sess.Insert(obj)
	return err
}
func (c *MySQLClient) UpdateByParamAnd(parmas []*ParamAnd, obj interface{}) error {
	switch len(parmas) {
	case 0:
		return fmt.Errorf("update failed. param lenth is 0")
	case 1:
		_, err := c.Where(fmt.Sprintf("%v %s ?", parmas[0].Col, parmas[0].Condition), parmas[0].Value).AllCols().Update(obj)
		return err
	default:
		sess := c.NewSession()
		defer sess.Close()
		_, err := sessionAnd(sess, parmas).AllCols().Update(obj)
		return err
	}
}
func UpdateBySessionParamAnd(sess *xorm.Session, parmas []*ParamAnd, obj interface{}) error {
	_, err := sessionAnd(sess, parmas).AllCols().Update(obj)
	return err
}
func (c *MySQLClient) DeleteByParamAnd(parmas []*ParamAnd, obj interface{}) error {
	switch len(parmas) {
	case 0:
		return fmt.Errorf("delete failed. param lenth is 0")
	case 1:
		_, err := c.Where(fmt.Sprintf("%v %s ?", parmas[0].Col, parmas[0].Condition), parmas[0].Value).Delete(obj)
		return err
	default:
		sess := c.NewSession()
		defer sess.Close()
		_, err := sessionAnd(sess, parmas).Delete(obj)
		return err
	}
}
func DeleteBySessionParamAnd(sess *xorm.Session, parmas []*ParamAnd, obj interface{}) error {
	_, err := sessionAnd(sess, parmas).Delete(obj)
	return err
}
func sessionAnd(sess *xorm.Session, parmas []*ParamAnd) *xorm.Session {
	for _, v := range parmas {
		sess = sess.Where(fmt.Sprintf("%v %s ?", v.Col, v.Condition), v.Value)
	}
	return sess
}
