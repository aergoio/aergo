function createTable()
    db.exec([[create table if not exists book (
          page number,
          contents text
      )]])
  
    db.exec([[create table if not exists copy_book (
          page number,
          contents text
      )]])
  end
  
  function makeBook()
         local stmt = db.prepare("insert into book values (?,?)")
      for i = 1, 100 do    
             stmt:exec(i, "value=" .. i*i)
      end
  end
  
  function copyBook()
      local rs = db.query("select page, contents from book order by page asc")
      while rs:next() do
          local col1, col2 = rs:get()
          local stmt_t = db.prepare("insert into copy_book values (?,?)")
          stmt_t:exec(col1, col2)
      end
  end
  
  
  function viewCopyBook()
      local rt = {}
      local rs = db.query("select max(page), min(contents) from copy_book")
      while rs:next() do
          local col1, col2 = rs:get()
          table.insert(rt, col1)
          table.insert(rt, col2)
      end
      return rt
  end
  
  function viewJoinBook()
      local rt = {}
      local rs = db.query([[select c.page, b.page, c.contents  
                              from copy_book c, book b 
                              where c.page = b.page and c.page = 10 ]])
      while rs:next() do
          local col1, col2, col3 = rs:get()
          table.insert(rt, col1)
          table.insert(rt, col2)
          table.insert(rt, col3)
      end
      return rt
  end
  
abi.register(createTable, makeBook, copyBook, viewCopyBook, viewJoinBook)