package contract

import "testing"

var testCases = map[string][2]bool{
	"PRAGMA index_info('idx52')":                                  {true, true},
	"/* PRAGMA */ insert into t values (1, 2)":                    {true, false},
	"insert /* PRAGMA */ into t values (1, 2)":                    {true, false},
	"insert/* PRAGMA */ into t values (1, 2)":                     {true, false},
	"/* pragma insert into t values (1, 2) */":                    {false, false},
	"-- pragma insert into t values (1, 2)":                       {false, false},
	"attach database test as \"test\"":                            {false, false},
	"attach*database test as \"test\"":                            {false, false},
	"'insert' into t values (1, 2)":                               {false, false},
	"select into t values (1, 2)":                                 {true, true},
	"create table t (a bigint, b text)":                           {true, false},
	"/* asdfasdf\n asdfadsf */ create table t (a bigint, b text)": {true, false},
	"-- asdfasdf\n asdfadsf create table t (a bigint, b text)":    {false, false},
	"-- asdfasdf\n create table t (a bigint, b text)":             {true, false},
	"insert\n-- asdfasdf\n create table t (a bigint, b text)":     {true, false},
	"create trigger x ...":                                        {false, false},
	"create view v ...":                                           {true, false},
	"create temp table tt ...":                                    {false, false},
	"create index":                                                {true, false},
	"/* blah -- blah ... */ create index":                         {true, false},
	"/* blah -- blah ... */ create unique index":                  {true, false},
	" create /* blah -- blah ... */unique index":                  {true, false},
	" create /* blah -- blah ... */unique table":                  {true, false},
}

func TestIsPermittedSql(t *testing.T) {
	for s, r := range testCases {
		expected := r[0]
		t.Log(s, expected)
		if cPermittedSql(s) != expected {
			t.Errorf("[FAIL] %s, expected: %v, got: %v\n", s, expected, !expected)
		}
	}
}
func TestIsReadOnlySql(t *testing.T) {
	for s, r := range testCases {
		expected := r[1]
		t.Log(s, expected)
		if cReadOnlySql(s) != expected {
			t.Errorf("[FAIL] %s, expected: %v, got: %v\n", s, expected, !expected)
		}
	}
}
