function createDataTypeTable()
    db.exec([[create table if not exists datatype_table(
          var1 varchar(10),
          char1 char(10),
          int1 int(5),
          float1 float(6),
          blockheight1 long
      )]])
  end
  
  function dropDataTypeTable()
     db.exec("drop table datatype_table")
  end
  
  function insertDataTypeTable()
      local stmt = db.prepare("insert into datatype_table values ('ABCD','fgh',1,3.14,?)")
      stmt:exec(system.getBlockheight())
  end
  function queryOrderByDesc()
      local rt = {}
      local rs = db.query("select * from datatype_table order by blockheight1 desc")
      while rs:next() do
          local col1, col2, col3, col4, col5 = rs:get()
          item = {
                      var1 = col1,
                      char1 = col2,
                      int1 = col3,
                      float1 = col4,
                      blockheight1 = col5
              }
          table.insert(rt, item)
      end
      return rt
  end
  
  function queryGroupByBlockheight1()
      local rt = {}
      local rs = db.query("select blockheight1, count(*), sum(int1), avg(float1) from datatype_table group by blockheight1")
      while rs:next() do
          local col1, col2, col3, col4 = rs:get()
          item = {
                      blockheight1 = col1,
                      count1 = col2,
                      sum_int1 = col3,
                      avg_float1 =col4
              }
          table.insert(rt, item)
      end
      return rt
  end
  
abi.register(createDataTypeTable, dropDataTypeTable, insertDataTypeTable, queryOrderByDesc, queryGroupByBlockheight1)