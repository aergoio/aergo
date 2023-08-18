//go:build Debug
// +build Debug

package contract

/*
#include "debug.h"
#include <stdlib.h>
*/
import "C"
import (
	"container/list"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/aergoio/aergo/v2/types"
)

type contract_info struct {
	contract_id_base58 string
	src_path           string
	breakpoints        *list.List
}

var contract_info_map = make(map[string]*contract_info)
var watchpoints = list.New()

func (ce *executor) setCountHook(limit C.int) {
	if ce == nil || ce.L == nil {
		return
	}
	if ce.err != nil {
		return
	}

	if cErrMsg := C.vm_set_debug_hook(ce.L); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)

		ctrLgr.Fatal().Str("err", errMsg).Msg("Fail to initialize lua contract debugger")
	}
}

func HexAddrToBase58Addr(contract_id_hex string) (string, error) {
	byteContractID, err := hex.DecodeString(contract_id_hex)
	if err != nil {
		return "", err
	}

	return types.NewAccount(byteContractID).ToString(), nil
}

func HexAddrOrPlainStrToHexAddr(d string) string {
	// try to convert to base58 address to test input format
	if encodedAddr, err := HexAddrToBase58Addr(d); err == nil {
		// try to decode the encoded str
		if _, err = types.DecodeAddress(encodedAddr); err == nil {
			// input is hex string. just return
			return d
		}
	}

	// input is a name to hashing or base58 address
	return PlainStrToHexAddr(d)
}

func PlainStrToHexAddr(d string) string {
	return hex.EncodeToString(StrHash(d))
}

func SetBreakPoint(contract_id_hex string, line uint64) error {

	if HasBreakPoint(contract_id_hex, line) {
		return errors.New("Same breakpoint already exists")
	}

	addr, err := HexAddrToBase58Addr(contract_id_hex)
	if err != nil {
		return err
	}

	if _, ok := contract_info_map[contract_id_hex]; !ok {
		// create new one if not exist
		contract_info_map[contract_id_hex] = &contract_info{
			addr,
			"",
			list.New()}
	}

	insertPoint := contract_info_map[contract_id_hex].breakpoints.Front()
	if insertPoint != nil {
		for {
			nextIter := insertPoint.Next()
			if line < insertPoint.Value.(uint64) {
				insertPoint = nil
				break
			} else if nextIter == nil || line < nextIter.Value.(uint64) {
				break
			}
			insertPoint = nextIter
		}
	}

	if insertPoint == nil {
		// line is the smallest or list is empty. insert the line to the first of the list
		contract_info_map[contract_id_hex].breakpoints.PushFront(line)
	} else {
		// insert after the most biggest breakpoints among smaller ones
		contract_info_map[contract_id_hex].breakpoints.InsertAfter(line, insertPoint)
	}

	return nil
}

func DelBreakPoint(contract_id_hex string, line uint64) error {
	if !HasBreakPoint(contract_id_hex, line) {
		return errors.New("Breakpoint does not exists")
	}

	if info, ok := contract_info_map[contract_id_hex]; ok {
		for iter := info.breakpoints.Front(); iter != nil; iter = iter.Next() {
			if line == iter.Value.(uint64) {
				info.breakpoints.Remove(iter)
				return nil
			}
		}
	}

	return nil
}

func HasBreakPoint(contract_id_hex string, line uint64) bool {
	if info, ok := contract_info_map[contract_id_hex]; ok {
		for iter := info.breakpoints.Front(); iter != nil; iter = iter.Next() {
			if line == iter.Value {
				return true
			}
		}
	}
	return false
}

//export PrintBreakPoints
func PrintBreakPoints() {
	if len(contract_info_map) == 0 {
		return
	}
	for _, info := range contract_info_map {
		fmt.Printf("%s (%s): ", info.contract_id_base58, info.src_path)
		for iter := info.breakpoints.Front(); iter != nil; iter = iter.Next() {
			fmt.Printf("%d ", iter.Value)
		}
		fmt.Printf("\n")
	}
}

//export ResetBreakPoints
func ResetBreakPoints() {
	for _, info := range contract_info_map {
		info.breakpoints = list.New()
	}
}

func SetWatchPoint(code string) error {
	if code == "" {
		return errors.New("Empty string cannot be set")
	}

	watchpoints.PushBack(code)

	return nil
}

func DelWatchPoint(idx uint64) error {
	if uint64(watchpoints.Len()) < idx {
		return errors.New("invalid index")
	}

	var i uint64 = 0
	for e := watchpoints.Front(); e != nil; e = e.Next() {
		i++
		if i >= idx {
			watchpoints.Remove(e)
			return nil
		}
	}

	return nil
}

func ListWatchPoints() *list.List {
	return watchpoints
}

//export ResetWatchPoints
func ResetWatchPoints() {
	watchpoints = list.New()
}

func UpdateContractInfo(contract_id_hex string, path string) {

	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			ctrLgr.Fatal().Str("path", path).Msg("Try to set a invalid path")
		}
		path = filepath.ToSlash(absPath)
	}

	if info, ok := contract_info_map[contract_id_hex]; ok {
		info.src_path = path

	} else {
		addr, err := HexAddrToBase58Addr(contract_id_hex)
		if err != nil {
			ctrLgr.Fatal().Str("contract_id_hex", contract_id_hex).Msg("Fail to Decode Hex Address")
		}
		contract_info_map[contract_id_hex] = &contract_info{
			addr,
			path,
			list.New()}
	}
}

func ResetContractInfo() {
	// just remove src paths. keep others for future use
	for _, info := range contract_info_map {
		info.src_path = ""
	}

}

//export CGetContractID
func CGetContractID(contract_id_hex_c *C.char) *C.char {
	contract_id_hex := C.GoString(contract_id_hex_c)
	if info, ok := contract_info_map[contract_id_hex]; ok {
		return C.CString(info.contract_id_base58)
	} else {
		return C.CString("")
	}
}

//export CGetSrc
func CGetSrc(contract_id_hex_c *C.char) *C.char {
	contract_id_hex := C.GoString(contract_id_hex_c)
	if info, ok := contract_info_map[contract_id_hex]; ok {
		return C.CString(info.src_path)
	} else {
		return C.CString("")
	}
}

//export CSetBreakPoint
func CSetBreakPoint(contract_name_or_hex_c *C.char, line_c C.double) {

	contract_name_or_hex := C.GoString(contract_name_or_hex_c)
	line := uint64(line_c)

	err := SetBreakPoint(HexAddrOrPlainStrToHexAddr(contract_name_or_hex), line)
	if err != nil {
		ctrLgr.Error().Err(err).Msg("Fail to add breakpoint")
	}
}

//export CDelBreakPoint
func CDelBreakPoint(contract_name_or_hex_c *C.char, line_c C.double) {
	contract_name_or_hex := C.GoString(contract_name_or_hex_c)
	line := uint64(line_c)

	err := DelBreakPoint(HexAddrOrPlainStrToHexAddr(contract_name_or_hex), line)
	if err != nil {
		ctrLgr.Error().Err(err).Msg("Fail to delete breakpoint")
	}
}

//export CHasBreakPoint
func CHasBreakPoint(contract_id_hex_c *C.char, line_c C.double) C.int {

	contract_id_hex := C.GoString(contract_id_hex_c)
	line := uint64(line_c)

	if HasBreakPoint(contract_id_hex, line) {
		return C.int(1)
	}

	return C.int(0)
}

//export CSetWatchPoint
func CSetWatchPoint(code_c *C.char) {
	code := C.GoString(code_c)

	err := SetWatchPoint(code)
	if err != nil {
		ctrLgr.Error().Err(err).Msg("Fail to set watchpoint")
	}
}

//export CDelWatchPoint
func CDelWatchPoint(idx_c C.double) {
	idx := uint64(idx_c)

	err := DelWatchPoint(idx)
	if err != nil {
		ctrLgr.Error().Err(err).Msg("Fail to del watchpoint")
	}
}

//export CGetWatchPoint
func CGetWatchPoint(idx_c C.int) *C.char {
	idx := int(idx_c)
	var i int = 0
	for e := watchpoints.Front(); e != nil; e = e.Next() {
		i++
		if i == idx {
			return C.CString(e.Value.(string))
		}
	}

	return C.CString("")
}

//export CLenWatchPoints
func CLenWatchPoints() C.int {
	return C.int(watchpoints.Len())
}

//export GetDebuggerCode
func GetDebuggerCode() *C.char {

	return C.CString(`
	package.preload['__debugger'] = function()

	--{{{  history

	--15/03/06 DCN Created based on RemDebug
	--28/04/06 DCN Update for Lua 5.1
	--01/06/06 DCN Fix command argument parsing
	--             Add step/over N facility
	--             Add trace lines facility
	--05/06/06 DCN Add trace call/return facility
	--06/06/06 DCN Make it behave when stepping through the creation of a coroutine
	--06/06/06 DCN Integrate the simple debugger into the main one
	--07/06/06 DCN Provide facility to step into coroutines
	--13/06/06 DCN Fix bug that caused the function environment to get corrupted with the global one
	--14/06/06 DCN Allow 'sloppy' file names when setting breakpoints
	--04/08/06 DCN Allow for no space after command name
	--11/08/06 DCN Use io.write not print
	--30/08/06 DCN Allow access to array elements in 'dump'
	--10/10/06 DCN Default to breakfile for all commands that require a filename and give '-'
	--06/12/06 DCN Allow for punctuation characters in DUMP variable names
	--03/01/07 DCN Add pause on/off facility
	--19/06/07 DCN Allow for duff commands being typed in the debugger (thanks to Michael.Bringmann@lsi.com)
	--             Allow for case sensitive file systems               (thanks to Michael.Bringmann@lsi.com)
	--04/08/09 DCN Add optional line count param to pause
	--05/08/09 DCN Reset the debug hook in Pause() even if we think we're started
	--30/09/09 DCN Re-jig to not use co-routines (makes debugging co-routines awkward)
	--01/10/09 DCN Add ability to break on reaching any line in a file
	--24/07/13 TWW Added code for emulating setfenv/getfenv in Lua 5.2 as per
	--             http://lua-users.org/lists/lua-l/2010-06/msg00313.html
	--25/07/13 TWW Copied Alex Parrill's fix for errors when tracing back across a C frame
	--             (https://github.com/ColonelThirtyTwo/clidebugger, 26/01/12)
	--25/07/13 DCN Allow for windows and unix file name conventions in has_breakpoint
	--26/07/13 DCN Allow for \ being interpreted as an escape inside a [] pattern in 5.2
	--29/01/17 RMM Fix lua 5.2 and 5.3 compat, fix crash in error msg, sort help output
	--22/03/19 Modified for Aergo Contracts

	--}}}
	--{{{  description

	--A simple command line debug system for Lua written by Dave Nichols of
	--Match-IT Limited. Its public domain software. Do with it as you wish.

	--This debugger was inspired by:
	-- RemDebug 1.0 Beta
	-- Copyright Kepler Project 2005 (http://www.keplerproject.org/remdebug)

	--Usage:
	--  require('debugger')        --load the debug library
	--  pause(message)             --start/resume a debug session

	--An assert() failure will also invoke the debugger.

	--}}}

	__debugger = {}

	local coro_debugger
	local events = { BREAK = 1, WATCH = 2, STEP = 3, SET = 4 }
	local watches = {}
	local step_into   = false
	local step_over   = false
	local step_lines  = 0
	local step_level  = {main=0}
	local stack_level = {main=0}
	local trace_level = {main=0}
	local ret_file, ret_line, ret_name
	local current_thread = 'main'
	local started = false
	local _g      = _G
	local skip_pause_for_init = false

	--{{{  make Lua 5.2 compatible

	local unpack = unpack or table.unpack
	local loadstring = loadstring or load

	--}}}

	--{{{  local hints -- command help
	--The format in here is name=summary|description
	local hints = {

		setb =    [[
	setb [line file]    -- set a breakpoint to line/file|, line 0 means 'any'

	If file is omitted or is '-'' the breakpoint is set at the file for the
	currently set level (see 'set'). Execution pauses when this line is about
	to be executed and the debugger session is re-activated.

	The file can be given as the fully qualified name, partially qualified or
	just the file name. E.g. if file is set as 'myfile.lua', then whenever
	execution reaches any file that ends with 'myfile.lua' it will pause. If
	no extension is given, any extension will do.

	If the line is given as 0, then reaching any line in the file will do.
	]],

	delb =    [[
	delb [line file]    -- removes a breakpoint|

	If file is omitted or is '-'' the breakpoint is removed for the file of the
	currently set level (see 'set').
	]],

	resetb = [[
	resetb             -- removes all breakpoints|
	]],

	setw =    [[
	setw <exp>          -- adds a new watch expression|

	The expression is evaluated before each line is executed. If the expression
	yields true then execution is paused and the debugger session re-activated.
	The expression is executed in the context of the line about to be executed.
	]],

	delw =    [[
	delw <index>        -- removes the watch expression at index|

	The index is that returned when the watch expression was set by setw.
	]],

	resetw = [[
	resetw             -- removes all watch expressions|
	]],

	run     = [[
	run                 -- run until next breakpoint or watch expression|
	]],

	step    = [[
	step [N]            -- run next N lines, stepping into function calls|

	If N is omitted, use 1.
	]],

	over    = [[
	over [N]            -- run next N lines, stepping over function calls|

	If N is omitted, use 1.
	]],

	out     = [[
	out [N]             -- run lines until stepped out of N functions|

	If N is omitted, use 1.
	If you are inside a function, using 'out 1' will run until you return
	from that function to the caller.
	]],

	listb   = [[
	listb               -- lists breakpoints|
	]],

	listw   = [[
	listw               -- lists watch expressions|
	]],

	set     = [[
	set [level]         -- set context to stack level, omitted=show|

	If level is omitted it just prints the current level set.
	This sets the current context to the level given. This affects the
	context used for several other functions (e.g. vars). The possible
	levels are those shown by trace.
	]],

	vars    = [[
	vars [depth]        -- list context locals to depth, omitted=1|

	If depth is omitted then uses 1.
	Use a depth of 0 for the maximum.
	Lists all non-nil local variables and all non-nil upvalues in the
	currently set context. For variables that are tables, lists all fields
	to the given depth.
	]],

	fenv    = [[
	fenv [depth]        -- list context function env to depth, omitted=1|

	If depth is omitted then uses 1.
	Use a depth of 0 for the maximum.
	Lists all function environment variables in the currently set context.
	For variables that are tables, lists all fields to the given depth.
	]],

	glob    = [[
	glob [depth]        -- list globals to depth, omitted=1|

	If depth is omitted then uses 1.
	Use a depth of 0 for the maximum.
	Lists all global variables.
	For variables that are tables, lists all fields to the given depth.
	]],

	ups     = [[
	ups                 -- list all the upvalue names|

	These names will also be in the 'vars' list unless their value is nil.
	This provides a means to identify which vars are upvalues and which are
	locals. If a name is both an upvalue and a local, the local value takes
	precedance.
	]],

	locs    = [[
	locs                -- list all the locals names|

	These names will also be in the 'vars' list unless their value is nil.
	This provides a means to identify which vars are upvalues and which are
	locals. If a name is both an upvalue and a local, the local value takes
	precedance.
	]],

	dump    = [[
	dump <var> [depth]  -- dump all fields of variable to depth|

	If depth is omitted then uses 1.
	Use a depth of 0 for the maximum.
	Prints the value of <var> in the currently set context level. If <var>
	is a table, lists all fields to the given depth. <var> can be just a
	name, or name.field or name.# to any depth, e.g. t.1.f accesses field
	'f' in array element 1 in table 't'.

	Can also be called from a script as dump(var,depth).
	]],

	trace   = [[
	trace               -- dumps a stack trace|

	Format is [level] = file,line,name
	The level is a candidate for use by the 'set' command.
	]],

	info    = [[
	info                -- dumps the complete debug info captured|

	Only useful as a diagnostic aid for the debugger itself. This information
	can be HUGE as it dumps all variables to the maximum depth, so be careful.
	]],

	show    = [[
	show line file X Y  -- show X lines before and Y after line in file|

	If line is omitted or is '-' then the current set context line is used.
	If file is omitted or is '-' then the current set context file is used.
	If file is not fully qualified and cannot be opened as specified, then
	a search for the file in the package[path] is performed using the usual
	'require' searching rules. If no file extension is given, .lua is used.
	Prints the lines from the source file around the given line.
	]],

	exit    = [[
	exit                -- exits debugger, re-start it using pause()|
	]],

	help    = [[
	help [command]      -- show this list or help for command|
	]],

	['<statement>'] = [[
	<statement>         -- execute a statement in the current context|

	The statement can be anything that is legal in the context, including
	assignments. Such assignments affect the context and will be in force
	immediately. Any results returned are printed. Use '=' as a short-hand
	for 'return', e.g. '=func(arg)' will call 'func' with 'arg' and print
	the results, and '=var' will just print the value of 'var'.
	]],

	}
	--}}}

	--{{{  local function getinfo(level,field)

	--like debug.getinfo but copes with no activation record at the given level
	--and knows how to get 'field'. 'field' can be the name of any of the
	--activation record fields or any of the 'what' names or nil for everything.
	--only valid when using the stack level to get info, not a function name.

	local function getinfo(level,field)
		level = level + 1  --to get to the same relative level as the caller
		if not field then return debug.getinfo(level) end
		local what
		if field == 'name' or field == 'namewhat' then
			what = 'n'
		elseif field == 'what' or field == 'source' or field == 'linedefined' or field == 'lastlinedefined' or field == 'short_src' then
			what = 'S'
		elseif field == 'currentline' then
			what = 'l'
		elseif field == 'nups' then
			what = 'u'
		elseif field == 'func' then
			what = 'f'
		else
			return debug.getinfo(level,field)
		end
		local ar = debug.getinfo(level,what)
		if ar then return ar[field] else return nil end
	end

	--}}}
	--{{{  local function indented( level, ... )

	local function indented( level, ... )
		io.write( string.rep('  ',level), table.concat({...}), '\n' )
	end

	--}}}
	--{{{  local function dumpval( level, name, value, limit )

	local function dumpval( level, name, value, limit )
		local index
		if type(name) == 'number' then
			index = string.format('[%d] = ',name)
		elseif type(name) == 'string'
			and (name == '__VARSLEVEL__' or name == '__ENVIRONMENT__' or name == '__GLOBALS__' or name == '__UPVALUES__' or name == '__LOCALS__') then
		--ignore these, they are debugger generated
			return
		elseif type(name) == 'string' and string.find(name,'^[_%a][_.%w]*$') then
			index = name ..' = '
		else
			index = string.format('[%q] = ',tostring(name))
		end
		if type(value) == 'table' then
			if (limit or 0) > 0 and level+1 >= limit then
				indented( level, index, tostring(value), ';' )
			else
				indented( level, index, '{' )
				for n,v in pairs(value) do
					dumpval( level+1, n, v, limit )
				end
				indented( level, '};' )
			end
		else
			if type(value) == 'string' then
				if string.len(value) > 40 then
					indented( level, index, '[[', value, ']];' )
				else
					indented( level, index, string.format('%q',value), ';' )
				end
			else
				indented( level, index, tostring(value), ';' )
			end
		end
	end

	--}}}
	--{{{  local function dumpvar( value, limit, name )

	local function dumpvar( value, limit, name )
		dumpval( 0, name or tostring(value), value, limit )
	end

	--}}}

	--{{{  local function show(contract_id_hex,line,before,after)

	--show +/-N lines of a contract source around line M

	local function show(contract_id_hex,line,before,after)

		line   = tonumber(line   or 1)
		before = tonumber(before or 10)
		after  = tonumber(after  or before)
		local file = ''
		local base58_addr = ''

		-- find matched source from 
		_, file = __get_contract_info(contract_id_hex)

		if not string.find(file,'%.') then file = file..'.lua' end

		local f = io.open(file,'r')
		if not f then
			io.write('Cannot find '..file..' for contract '..base58_addr..'\n')
		return
		end

		local i = 0
		for l in f:lines() do
			i = i + 1
			if i >= (line-before) then
				if i > (line+after) then break end
				if i == line then
					io.write(i..'***\t'..l..'\n')
				else
					io.write(i..'\t'..l..'\n')
				end
			end
		end

		f:close()

	end

	--}}}
	--{{{  local function tracestack(l)

	local function gi( i )
		return function() i=i+1 return debug.getinfo(i),i end
	end

	local function gl( level, j )
		return function() j=j+1 return debug.getlocal( level, j ) end
	end

	local function gu( func, k )
		return function() k=k+1 return debug.getupvalue( func, k ) end
	end

	local  traceinfo

	local function tracestack(l)
		local l = l + 1                        --NB: +1 to get level relative to caller
		traceinfo = {}
		--traceinfo.pausemsg = pausemsg
		for ar,i in gi(l) do
			table.insert( traceinfo, ar )
			if ar.what ~= 'C' then
				local names  = {}
				local values = {}
				
				for n,v in gl(i-1,0) do
				--for n,v in gl(i,0) do
				if string.sub(n,1,1) ~= '(' then   --ignore internal control variables
					table.insert( names, n )
					table.insert( values, v )
				end
				end
				if #names > 0 then
					ar.lnames  = names
					ar.lvalues = values
				end
			end
			if ar.func then
				local names  = {}
				local values = {}
				for n,v in gu(ar.func,0) do
				if string.sub(n,1,1) ~= '(' then   --ignore internal control variables
					table.insert( names, n )
					table.insert( values, v )
				end
				end
				if #names > 0 then
					ar.unames  = names
					ar.uvalues = values
				end
			end
		end
	end

	--}}}
	--{{{  local function trace()

	local function trace(set)
		local mark
		for level,ar in ipairs(traceinfo) do
			if level == set then
				mark = '***'
			else
				mark = ''
			end
			local contract_id_base58, _ = __get_contract_info(ar.source)
			io.write('['..level..']'..mark..'\t'..(ar.name or ar.what)..' in '..(contract_id_base58 or ar.short_src)..':'..ar.currentline..'\n')
		end
	end

	--}}}
	--{{{  local function info()

	local function info() dumpvar( traceinfo, 0, 'traceinfo' ) end

	--}}}

	--}}}
	--{{{  local function has_breakpoint(file, line)

	--allow for 'sloppy' file names
	--search for file and all variations walking up its directory hierachy
	--ditto for the file with no extension
	--a breakpoint can be permenant or once only, if once only its removed
	--after detection here, these are used for temporary breakpoints in the
	--a breakpoint on line 0 of a file means any line in that file

	local function has_breakpoint(contract_id_hex, line)
		
		return __has_breakpoint(contract_id_hex, line)
	end

	--}}}
	--{{{  local function capture_vars(ref,level,line)

	local function capture_vars(ref,level,line)
		--get vars, contract_id_hex, contract_id_base58 and line for the given level relative to debug_hook offset by ref

		local lvl = ref + level                --NB: This includes an offset of +1 for the call to here

		--{{{  capture variables

		local ar = debug.getinfo(lvl, 'f')
		if not ar then return {},'?','?',0 end

		local vars = {__UPVALUES__={}, __LOCALS__={}}
		local i

		local func = ar.func
		if func then
			i = 1
			while true do
				local name, value = debug.getupvalue(func, i)
				if not name then break end
				if string.sub(name,1,1) ~= '(' then  --NB: ignoring internal control variables
					vars[name] = value
					vars.__UPVALUES__[i] = name
				end
				i = i + 1
			end
			vars.__ENVIRONMENT__ = getfenv(func)
		end

		vars.__GLOBALS__ = getfenv(0)

		i = 1
		while true do
			local name, value = debug.getlocal(lvl, i)
			if not name then break end
			if string.sub(name,1,1) ~= '(' then    --NB: ignoring internal control variables
				vars[name] = value
				vars.__LOCALS__[i] = name
			end
			i = i + 1
		end

		vars.__VARSLEVEL__ = level

		if func then
			--NB: Do not do this until finished filling the vars table
			setmetatable(vars, { __index = getfenv(func), __newindex = getfenv(func) })
		end

		--NB: Do not read or write the vars table anymore else the metatable functions will get invoked!

		--}}}

		local contract_id_hex = getinfo(lvl, 'source')
		if string.find(contract_id_hex, '@') == 1 then
			contract_id_hex = string.sub(contract_id_hex, 2)
		end

		local contract_id_base58, _ = __get_contract_info(contract_id_hex)
		
		if not line then
			line = getinfo(lvl, 'currentline')
		end

		return vars,contract_id_hex,contract_id_base58,line

	end

	--}}}
	--{{{  local function restore_vars(ref,vars)

	local function restore_vars(ref,vars)

		if type(vars) ~= 'table' then return end

		local level = vars.__VARSLEVEL__       --NB: This level is relative to debug_hook offset by ref
		if not level then return end

		level = level + ref                    --NB: This includes an offset of +1 for the call to here

		local i
		local written_vars = {}

		i = 1
		while true do
			local name, value = debug.getlocal(level, i)
			if not name then break end
			if vars[name] and string.sub(name,1,1) ~= '(' then     --NB: ignoring internal control variables
				debug.setlocal(level, i, vars[name])
				written_vars[name] = true
			end
			i = i + 1
		end

		local ar = debug.getinfo(level, 'f')
		if not ar then return end

		local func = ar.func
		if func then

		i = 1
		while true do
			local name, value = debug.getupvalue(func, i)
			if not name then break end
			if vars[name] and string.sub(name,1,1) ~= '(' then   --NB: ignoring internal control variables
				if not written_vars[name] then
					debug.setupvalue(func, i, vars[name])
				end
				written_vars[name] = true
			end
			i = i + 1
		end

		end

	end

	--}}}
	--{{{  local function trace_event(event, line, level)

		local function trace_event(event, line, level)

		if event ~= 'line' then return end

		local slevel = stack_level[current_thread]
		local tlevel = trace_level[current_thread]

		trace_level[current_thread] = stack_level[current_thread]

	end

	--}}}
	--{{{  local function report(ev, vars, file, line, idx_watch)

	local function report(ev, vars, contract_id_base58, line, idx_watch)
		local vars = vars or {}
		local contract_id_base58 = contract_id_base58 or '?'
		local line = line or 0
		local prefix = ''
		if current_thread ~= 'main' then prefix = '['..tostring(current_thread)..'] ' end
		if ev == events.STEP then
			io.write(prefix..'Paused at contract '..contract_id_base58..' line '..line..' ('..stack_level[current_thread]..')\n')
		elseif ev == events.BREAK then
			io.write(prefix..'Paused at contract '..contract_id_base58..' line '..line..' ('..stack_level[current_thread]..') (breakpoint)\n')
		elseif ev == events.WATCH then
			io.write(prefix..'Paused at contract '..contract_id_base58..' line '..line..' ('..stack_level[current_thread]..')'..' (watch expression '..idx_watch.. ': ['..__get_watchpoint(idx_watch)..'])\n')
		elseif ev == events.SET then
			--do nothing
		else
			io.write(prefix..'Error in application: '..contract_id_base58..' line '..line..'\n')
		end
		return vars, contract_id_base58, line
	end

	--}}}

	--{{{  local function debugger_loop(ev, vars, file, line, idx_watch)

	local function debugger_loop(ev, vars, file, line, idx_watch)

		local eval_env  = vars or {}
		local breakfile = file or '?'
		local breakline = line or 0

		local command, args

		--{{{  local function getargs(spec)

		--get command arguments according to the given spec from the args string
		--the spec has a single character for each argument, arguments are separated
		--by white space, the spec characters can be one of:
		-- F for a filename    (defaults to breakfile if - given in args)
		-- L for a line number (defaults to breakline if - given in args)
		-- N for a number
		-- V for a variable name
		-- S for a string

		local function getargs(spec)
			local res={}
			local char,arg
			local ptr=1
			for i=1,string.len(spec) do
				char = string.sub(spec,i,i)
				if     char == 'F' then
					_,ptr,arg = string.find(args..' ','%s*([%w%p]*)%s*',ptr)
					if not arg or arg == '' then arg = '-' end
					if arg == '-' then arg = breakfile end
				elseif char == 'L' then
					_,ptr,arg = string.find(args..' ','%s*([%w%p]*)%s*',ptr)
					if not arg or arg == '' then arg = '-' end
					if arg == '-' then arg = breakline end
					arg = tonumber(arg) or 0
				elseif char == 'N' then
					_,ptr,arg = string.find(args..' ','%s*([%w%p]*)%s*',ptr)
					if not arg or arg == '' then arg = '0' end
					arg = tonumber(arg) or 0
				elseif char == 'V' then
					_,ptr,arg = string.find(args..' ','%s*([%w%p]*)%s*',ptr)
					if not arg or arg == '' then arg = '' end
				elseif char == 'S' then
					_,ptr,arg = string.find(args..' ','%s*([%w%p]*)%s*',ptr)
					if not arg or arg == '' then arg = '' end
				else
					arg = ''
				end
				table.insert(res,arg or '')
			end
			return unpack(res)
		end

		--}}}

		while true do
			io.write('[DEBUG]> ')
			local line = io.read('*line')
			if line == nil then io.write('\n'); line = 'exit' end

			if string.find(line, '^[a-z]+') then
				command = string.sub(line, string.find(line, '^[a-z]+'))
				args    = string.gsub(line,'^[a-z]+%s*','',1)            --strip command off line
			else
				command = ''
			end

			if command == 'setb' then
				--{{{  set breakpoint

				local line, contract_id_hex  = getargs('LF')
				if contract_id_hex ~= '' and line ~= '' then
					__set_breakpoint(contract_id_hex,line)
				else
					io.write('Bad request\n')
				end

				--}}}

			elseif command == 'delb' then
				--{{{  delete breakpoint

				local line, contract_id_hex = getargs('LF')
				if contract_id_hex ~= '' and line ~= '' then
					__delete_breakpoint(contract_id_hex, line)
				else
					io.write('Bad request\n')
				end

				--}}}

			elseif command == 'resetb' then
				--{{{  delete all breakpoints
				--TODO
				io.write('All breakpoints deleted\n')
				--}}}

			elseif command == 'listb' then
				--{{{  list breakpoints
				__print_breakpoints()
				--}}}

			elseif command == 'setw' then
				--{{{  set watch expression

				if args and args ~= '' then
					__set_watchpoint(args)
					io.write('Set watch exp no. ' .. __len_watchpoints() ..'\n')
				else
					io.write('Bad request\n')
				end

				--}}}

			elseif command == 'delw' then
				--{{{  delete watch expression

				local index = tonumber(args)
				if index then
					__delete_watchpoint(index)
					io.write('Watch expression deleted\n')
				else
					io.write('Bad request\n')
				end

				--}}}

			elseif command == 'resetw' then
				--{{{  delete all watch expressions
				__reset_watchpoints()
				io.write('All watch expressions deleted\n')
				--}}}

			elseif command == 'listw' then
				--{{{  list watch expressions
				for i, v in pairs(__list_watchpoints()) do
					io.write(i .. ': ' .. v..'\n')
				end
				--}}}

			elseif command == 'run' then
				--{{{  run until breakpoint
				step_into = false
				step_over = false
				return 'cont'
				--}}}

			elseif command == 'step' then
				--{{{  step N lines (into functions)
				local N = tonumber(args) or 1
				step_over  = false
				step_into  = true
				step_lines = tonumber(N or 1)
				return 'cont'
				--}}}

			elseif command == 'over' then
				--{{{  step N lines (over functions)
				local N = tonumber(args) or 1
				step_into  = false
				step_over  = true
				step_lines = tonumber(N or 1)
				step_level[current_thread] = stack_level[current_thread]
				return 'cont'
				--}}}

			elseif command == 'out' then
				--{{{  step N lines (out of functions)
				local N = tonumber(args) or 1
				step_into  = false
				step_over  = true
				step_lines = 1
				step_level[current_thread] = stack_level[current_thread] - tonumber(N or 1)
				return 'cont'
				--}}}

			elseif command == 'set' then
				--{{{  set/show context level
				local level = args
				if level and level == '' then level = nil end
				if level then return level end
				--}}}

			elseif command == 'vars' then
				--{{{  list context variables
				local depth = args
				if depth and depth == '' then depth = nil end
				depth = tonumber(depth) or 1
				dumpvar(eval_env, depth+1, 'variables')
				--}}}

			elseif command == 'glob' then
				--{{{  list global variables
				local depth = args
				if depth and depth == '' then depth = nil end
				depth = tonumber(depth) or 1
				dumpvar(eval_env.__GLOBALS__,depth+1,'globals')
				--}}}

			elseif command == 'fenv' then
				--{{{  list function environment variables
				local depth = args
				if depth and depth == '' then depth = nil end
				depth = tonumber(depth) or 1
				dumpvar(eval_env.__ENVIRONMENT__,depth+1,'environment')
				--}}}

			elseif command == 'ups' then
				--{{{  list upvalue names
				dumpvar(eval_env.__UPVALUES__,2,'upvalues')
				--}}}

			elseif command == 'locs' then
				--{{{  list locals names
				dumpvar(eval_env.__LOCALS__,2,'upvalues')
				--}}}

			elseif command == 'dump' then
				--{{{  dump a variable
				local name, depth = getargs('VN')
				if name ~= '' then
					if depth == '' or depth == 0 then depth = nil end
					depth = tonumber(depth or 1)
					local v = eval_env
					local n = nil
					for w in string.gmatch(name,'[^%.]+') do     --get everything between dots
						if tonumber(w) then
							v = v[tonumber(w)]
						else
							v = v[w]
						end
						if n then n = n..'.'..w else n = w end
						if not v then break end
					end
					dumpvar(v,depth+1,n)
				else
					io.write('Bad request\n')
				end
				--}}}

			elseif command == 'show' then
				--{{{  show contract around a line or the current breakpoint
				local line, contract_id_hex, before, after = getargs('LFNN')
				if before == 0 then before = 10     end
				if after  == 0 then after  = before end

				if contract_id_hex ~= '' and contract_id_hex ~= '=stdin' then
					show(contract_id_hex,line,before,after)
				else
					io.write('Nothing to show\n')
				end
				--}}}

			elseif command == 'trace' then
				--{{{  dump a stack trace
				trace(eval_env.__VARSLEVEL__)
				--}}}

			elseif command == 'info' then
				--{{{  dump all debug info captured
				info()
				--}}}

			elseif command == 'pause' then
				--{{{  not allowed in here
				io.write('pause() should only be used in the script you are debugging\n')
				--}}}

			elseif command == 'help' then
				--{{{  help
				local command = getargs('S')
				if command ~= '' and hints[command] then
					io.write(hints[command]..'\n')
				else
					local l = {}
					for k,v in pairs(hints) do
						local _,_,h = string.find(v,'(.+)|')
						l[#l+1] = h..'\n'
					end
					table.sort(l)
					io.write(table.concat(l))
				end
				--}}}

			elseif command == 'exit' then
				--{{{  exit debugger
				return 'stop'
				--}}}

			elseif line ~= '' then
				--{{{  just execute whatever it is in the current context

				--map line starting with '=...' to 'return ...'
				if string.sub(line,1,1) == '=' then line = string.gsub(line,'=','return ',1) end

				local ok, func = pcall(loadstring,line)
				if ok and func==nil then -- auto-print variables
					ok, func = pcall(loadstring,'io.write(tostring(' .. line .. '))')
				end
				if func == nil then                             --Michael.Bringmann@lsi.com
					io.write('Compile error: '..line..'\n')
				elseif not ok then
					io.write('Compile error: '..func..'\n')
				else
					setfenv(func, eval_env)
					local res = {pcall(func)}
					if res[1] then
						if res[2] then
						table.remove(res,1)
						for _,v in ipairs(res) do
							io.write(tostring(v))
							io.write('\t')
						end
						end
						--update in the context
						io.write('\n')
						return 0
					else
						io.write('Run error: '..res[2]..'\n')
					end
				end

				--}}}
			end
		end

	end

	--}}}
	--{{{  local function debug_hook(event, line, level, thread)
	local function debug_hook(event, line, level, thread)
		if not started then debug.sethook(); coro_debugger = nil; return end
		current_thread = thread or 'main'
		local level = level or 2
		trace_event(event,line,level)
		if event == 'line' then
			-- calculate current stack
			for i=1,99999,1 do 
				if not debug.getinfo(i) then break end
				stack_level[current_thread] = i - 1 -- minus one to remove this debug_hook stack
			end
			
			local vars,contract_id_hex,contract_id_base58,line = capture_vars(level,1,line)
			local stop, ev, idx = false, events.STEP, 0
			while true do
				for index, value in pairs(__list_watchpoints()) do
				local func = loadstring('return(' .. value .. ')')
				if func ~= nil then
					setfenv(func, vars)
					local status, res = pcall(func)
					if status and res then
						ev, idx = events.WATCH, index
						stop = true
						break
					end
				end
				end
				if stop then break end
				if (step_into)
				or (step_over and (stack_level[current_thread] <= step_level[current_thread] or stack_level[current_thread] == 0)) then
					step_lines = step_lines - 1
					if step_lines < 1 then
						ev, idx = events.STEP, 0
						break
					end
				end
				if has_breakpoint(contract_id_hex, line) then
					ev, idx = events.BREAK, 0
					break
				end
				return
			end
			if skip_pause_for_init then
				--DO notthing
			elseif not coro_debugger then
				io.write('Lua Debugger\n')
				vars, contract_id_base58, line = report(ev, vars, contract_id_base58, line, idx)
				io.write('Type \'help\' for commands\n')
				coro_debugger = true
			else
				vars, contract_id_base58, line = report(ev, vars, contract_id_base58, line, idx)
			end
			tracestack(level)
			local last_next = 1
			local next = 'ask'
			local silent = false
			while true do
				if next == 'ask' then
					if skip_pause_for_init then 
						step_into = false
						--step_over = false
						skip_pause_for_init = false -- reset flag
						return -- for the first time
					end
					next = debugger_loop(ev, vars, contract_id_hex, line, idx)
				elseif next == 'cont' then
					return
				elseif next == 'stop' then
					started = false
					debug.sethook()
					coro_debugger = nil
					return
				elseif tonumber(next) then --get vars for given level or last level
					next = tonumber(next)
					if next == 0 then silent = true; next = last_next else silent = false end
					last_next = next
					restore_vars(level,vars)
					vars, contract_id_hex, contract_id_base58, line = capture_vars(level,next)
					if not silent then
						if vars and vars.__VARSLEVEL__ then
							io.write('Level: '..vars.__VARSLEVEL__..'\n')
						else
							io.write('No level set\n')
						end
					end
					ev = events.SET
					next = 'ask'
				else
					io.write('Unknown command from debugger_loop: '..tostring(next)..'\n')
					io.write('Stopping debugger\n')
					next = 'stop'
				end
			end
		end
	end


	--{{{  function hook()

	--
	-- Init and Register Debug Hook
	--
	--function hook()

	function __debugger.hook()
		--set to stop when get out of pause()
		trace_level[current_thread] = 0
		step_level [current_thread] = 0
		stack_level[current_thread] = 1
		step_lines = 1
		step_into = true
		started    = true
		skip_pause_for_init = true

		debug.sethook(debug_hook, 'l')   
	end

	--}}}
	--{{{  function dump(v,depth)

	--shows the value of the given variable, only really useful
	--when the variable is a table
	--see dump debug command hints for full semantics

	local function dump(v,depth)
		dumpvar(v,(depth or 1)+1,tostring(v))
	end

	--}}}

	return __debugger
end

require('__debugger').hook()
`)
}
