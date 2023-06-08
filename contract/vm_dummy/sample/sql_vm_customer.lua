function createTable()
    db.exec([[create table if not exists customer(
          id varchar(10),
          passwd varchar(20),
          name varchar(30),
          birth char(8),
          mobile varchar(20)
      )]])
  end
  
  function query(id)
      local rt = {}
      local rs = db.query("select * from customer where id like '%' || ? || '%'", id)
      while rs:next() do
          local col1, col2, col3, col4, col5 = rs:get()
          local item = {
                      id = col1,
                      passwd = col2,
                      name = col3,
                      birth = col4,
                      mobile = col5
              }
          table.insert(rt, item)
      end
      return rt
  end
  
  function insert(id , passwd, name, birth, mobile)
      local n = db.exec("insert into customer values (?,?,?,?,?)", id, passwd, name, birth, mobile)
      assert(n == 1, "insert count mismatch")
  end
  
  function update(id , passwd)
      local n = db.exec("update customer set passwd =? where id =?", passwd, id)
      assert(n == 1, "update count mismatch")
  end
  
  function delete(id)
      local n = db.exec("delete from customer where id =?", id)
      assert(n == 1, "delete count mismatch")
  end
  
  function count()
      local rs = db.query("select count(*) from customer")
      if rs:next() then
          local n = rs:get()
          return n
      else
          return "error in count()"
      end
  end
  
abi.register(createTable, query, insert, update, delete, count)