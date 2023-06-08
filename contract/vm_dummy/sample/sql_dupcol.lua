function get()
	db.query("select * from (select 1+1, 1+1, 1+1, 1+1, 1+1, 1+1)")
	return "success"
end

abi.register(get)