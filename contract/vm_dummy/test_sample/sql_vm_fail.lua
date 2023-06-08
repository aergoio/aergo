function init()
    db.exec("create table if not exists total(n int)")
	db.exec("insert into total values (0)")
end

function add(n)
	local stmt = db.prepare("update total set n = n + ?")
	stmt:exec(n)
end

function addFail(n)
	local stmt = db.prepare("update set n = n + ?")
	stmt:exec(n)
end

function get()
	local rs = db.query("select n from total")
	rs:next()
	n = rs:get()
	return n
end
abi.register(init, add, addFail, get)