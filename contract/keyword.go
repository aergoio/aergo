package contract

/*
#include <stdlib.h>
#include "sqlcheck.h"
*/
import "C"
import (
	"unsafe"

	"github.com/derekparker/trie"
)

var keywords *trie.Trie

func init() {
	keywords = trie.New()
	keywords.Add("ALTER", nil)
	keywords.Add("CREATE", nil)
	keywords.Add("DELETE", nil)
	keywords.Add("DROP", nil)
	keywords.Add("INSERT", nil)
	keywords.Add("REINDEX", nil)
	keywords.Add("REPLACE", nil)
	keywords.Add("SELECT", nil)
	keywords.Add("UPDATE", nil)
}

//export PermittedCmd
func PermittedCmd(cmd *C.char) C.int {
	if _, ok := keywords.Find(C.GoString(cmd)); ok {
		return C.int(1)
	}
	return C.int(0)
}

func cPermittedSql(sql string) bool {
	cstr := C.CString(sql)
	r := C.sqlcheck_is_permitted_sql(cstr)
	var b bool
	if r == C.int(1) {
		b = true
	}
	C.free(unsafe.Pointer(cstr))
	return b
}

func cReadOnlySql(sql string) bool {
	cstr := C.CString(sql)
	r := C.sqlcheck_is_readonly_sql(cstr)
	var b bool
	if r == C.int(1) {
		b = true
	}
	C.free(unsafe.Pointer(cstr))
	return b
}
