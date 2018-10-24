ASCL Reference (Draft)
==============

[TOC]

## Introduction

ASCL(Aergo Smart Contract Lanaguage)는 Aergo를 위한 Smart Contract 전용 언어다. ASCL은 strongly typed 언어로서 C, Java, Go 등과 같은 일반적인 언어 구문과 database 접근을 위한 SQL 구문을 제공한다.

### Notation

이 문서에서 각 문법은 [Extended Backus-Naur Form (EBNF)](https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form)를 사용하여 표기한다.

### Source code representations

ASCL로 작성된 프로그램은 다음 사항을 준수해야 한다.

* [UTF-8](https://en.wikipedia.org/wiki/UTF-8) 형식으로 작성해야 한다.
* 대소문자를 구분한다.

## Lexical elements

### Letters and digits

<pre>
<a name="letter">letter</a>        = ? [a-zA-Z_] ? ;
<a name="unicode">unicode</a>       = ? 임의의 unicode 범위 문자 ? ;

<a name="decimal_digit">decimal_digit</a> = ? [0-9] ? ;
<a name="octal_digit">octal_digit</a>   = ? [0-7] ? ;
<a name="hex_digit">hex_digit</a>     = ? [0-9a-fA-F] ? ;
</pre>

### Comments

comments는 다음과 같은 두가지 형식을 지원한다.

* Single-line comments (// 으로 시작하여 그 줄의 끝까지)

```
// This is a single-line comment
```

* Multi-line comments (/* 으로 시작하여 */ 까지)

```
/*
 * This is a multi-line comment
 */
```

### Constants

상수는 프로그램에서 절대 변하지 않는 값을 나타내며, integer, floating-point, string, boolean 상수등이 있다.

#### Boolean literals

boolean constant는 predefined constant로 true, false를 사용할 수 있다.

<pre>
<a name="bool">bool</a> = "true" | "false"
</pre>

#### Integer literals

integer constant는 10진수, 8진수, 16진수 형태로 표기할 수 있다.

<pre>
<a name="integer">integer</a>     = <a href="#decimal">decimal</a> | <a href="#octal">octal</a> | <a href="#hexadecimal">hexadecimal</a> ;
<a name="decimal">decimal</a>     = ( <a href="#decimal_digit">decimal_digit</a> - "0" ) { <a href="#decimal_digit">decimal_digit</a>  } ;
<a name="octal">octal</a>       = "0" { <a href="#octal_digit">octal_digit</a> } ;
<a name="hexadecimal">hexadecimal</a> = "0" ( "x" | "X" ) <a href="#hex_digit">hex_digit</a> { <a href="#hex_digit">hex_digit</a> } ;
</pre>

```
10
0700
0xff
```

#### Floating-point literals

floating-point constant는 소수점을 사용하거나 exponent를 사용하여 표기할 수 있다.

<pre>
<a name="float">float</a>    = <a href="#natural">natural</a> "." [ <a href="#natural">natural</a> ] [ <a href="#exponent">exponent</a> ] |
           "." <a href="#natural">natural</a> [ <a href="#exponent">exponent</a> ] |
           <a href="#natural">natural</a> <a href="#exponent">exponent</a> ;
<a name="natural">natural</a>  = <a href="#decimal_digit">decimal_digit</a> { <a href="#decimal_digit">decimal_digit</a> } ;
<a name="exponent">exponent</a> = ( "e" | "E" ) [ "+" | "-" ] { <a href="#natural">natural</a> } ;
</pre>

```
1.23
71.
.009
1E3
1e+19
3.14e-02
.4E+3
```

#### String literals

string constant는 쌍따옴표를 사용하여 표기할 수 있다.

<pre>
<a name="string">string</a> = "\"" { <a href="#unicode">unicode</a> } "\"" ;
</pre>

### Identifiers

identifier는 변수나 타입의 이름으로 사용되며, 첫글자는 반드시 문자여야 한다.

<pre>
<a name="id">identifier</a> = <a href="#letter">letter</a> { <a href="#letter">letter</a> | <a href="#decimal_digit">decimal_digit</a> } ;
</pre>

```
abc
Var1
_case
```

### Keywords

다음은 ASCL에서 사용하는 기본 예약어이고, identifier로 사용할 수 없다.

```
break       case        const       continue    contract    default
else        enum        false       for         func        goto
if          import      in          new         null        payable
public      readonly    return      struct      switch      true
```

다음은 ASCL에서 사용하는 SQL 예약어이고, 마찬가지로 identifier로 사용할 수 없다.

```
create      delete      drop        insert      select      update
```

### Operators and punctuations

다음은 ASCL에서 사용할 수 있는 operator다.

```
+           -           *           /           %           =
+=          -=          *=          /=          %=          &
|           ^           &=          |=          ^=          >>=
<<=         >>          <<          &&          ||          <
>           <=          >=          ==          !=          ++
--          ?           :
```

다음은 ASCL에서 사용할 수 있는 punctuation이다.

```
(           )           {           }           [           ]
;           ,           .           !
```

### Types

<pre>
<a name="type">Type</a>        = <a href="#prim_type">PrimType</a> | <a href="#complex_type">ComplexType</a> ;
<a name="prim_type">PrimType</a>    = "bool" | "byte" | "int8" | "int16" | "int32" | "int64" | "uint8" |
              "uint16" | "uint32" | "uint64" | "float" | "double" | "string" ;
<a name="complex_type">ComplexType</a> = <a href="#struct_decl">StructDecl</a> | <a href="#enum_decl">EnumDecl</a> | <a href="#map_decl">MapDecl</a> ;
</pre>

#### Boolean types

boolean type은 true와 false 값을 저장할 수 있는 타입으로 bool 타입을 사용할 수 있다.

```
bool v = true
bool v = false
```

#### Numeric types

numeric type은 integer 타입과 floating-point 타입을 사용할 수 있다.

```
int8        8-bit integers (-128 ~ 127)
int16       16-bit integers (-32768 ~ 32767)
int32       32-bit integers (-2147483648 ~ 2147483647)
int64       64-bit integers (-9223372036854775808 ~ 9223372036854775807)

uint8       8-bit integers (0 ~ 255)
uint16      16-bit integers (0 ~ 65535)
uint32      32-bit integers (0 ~ 4294967295)
uint64      64-bit integers (0 ~ 18446744073709551615)

float       IEEE-754 32-bit floating-point numbers
double      IEEE-754 64-bit floating-point numbers

byte        alias for uint8
int         alias for int32
uint        alias for uint32
```

```
int8 v = 0
uint32 v = 2018
```

#### String types

string type은 문자열을 저장할 수 있는 타입으로 string 타입을 사용할 수 있다.

```
string v = "This is a string"
string v = "This is a
    string too";
```

#### Struct types

struct type은 1개 이상의 field들의 집합으로 다음과 같이 선언한다. struct type은 중복하여 사용할 수 있다.

<pre>
<a name="struct_decl">StructDecl</a> = "struct" <a href="#id">identifier</a> "{" { <a href="#field_decl">FieldDecl</a> ";" } "}" ;
<a name="field_decl">FieldDecl</a>  = <a href="#type">Type</a> <a href="#id">identifier</a> ;
</pre>

```
struct User {
    int id;
    string name;
    string address;
}
struct Customer {
    User user;
    int purchase_count;
}
```

#### Enum types

enum type은 1개 이상의 constant element들의 집합으로 다음과 같이 선언한다. 만약 별도의 값이 정의되지 않을 경우 첫번째 element부터 0이 저장되고, 차례로 1씩 증가된 값이 저장된다.

<pre>
<a name="enum_decl">EnumDecl</a>       = "enum" <a href="#id">identifier</a> "{" { <a href="#enum_decl">EnumeratorDecl</a> "," } "}" ;
<a name="enum_decl">EnumeratorDecl</a> = <a href="#id">identifier</a> [ "=" <a href="#decimal_digit">decimal_digit</a> ] ;
</pre>

```
enum Medal {
    GOLD,           // == 0
    SILVER,         // == 1
    BRONZE,         // == 2
}
enum City {
    Seoul,          // == 0
    Incheon,        // == 1
    Busan = 5,      // == 5
    Jeju,           // == 6
}
```

#### Map types

map type은 key, value를 쌍으로 갖는 hashmap 타입이다. map에서 사용하는 key 타입은 반드시 비교가능한 타입이어야 하나, value 타입은 임의의 타입을 사용할 수 있다.

<pre>
<a name="map_decl">MapDecl</a> = "map" "(" <a href="#type">Type</a> "," <a href="#type">Type</a> ")" ;
</pre>

```
map(int, string)
map(int, User)
map(int, map(string, string))
```

## Program structures

Smart Contract는 다음과 같이 import 선언부와 contract 선언부로 나눠진다.

<pre>
SmartContract = { <a href="#import_decl">ImportDecl</a> } { <a href="#contract_decl">ContractDecl</a> } ;
</pre>

### Import declarations

import는 외부 smart contract 파일을 참조하기 위해 사용하는 구문이다. import하기 위한 파일은 local storage만 참조할 수 있고, 절대경로를 사용하거나 현재 contract 파일이 존재하는 위치로부터의 상대경로를 사용할 수 있다.

<pre>
<a name="import_decl">ImportDecl</a> = "import" <a href="#string">string</a>
</pre>

```
import "ExternalContract"
import "../RootContract"
import "/home/CommonContract"
```

### Contract declarations

contract는 실제 로직을 작성하는 부분으로 다음과 같이 선언한다.

<pre>
<a name="contract_decl">ContractDecl</a> = "contract" <a href="#id">identifier</a> "{" <a href="#contract_body">ContractBody</a> "}" ;
<a name="contract_body">ContractBody</a> = { <a href="#var_decl">VariableDecl</a> | <a href="#struct_decl">StructDecl</a> | <a href="#func_decl">FunctionDecl</a> | <a href="#ctor_decl">ConstructorDecl</a> } ;
</pre>

```
contract Bank { ... }
```

### Variables

variable은 값을 저장하기 위한 공간으로 다음과 같이 선언한다.

<pre>
<a name="var_decl">VariableDecl</a>   = [ <a href="#qual">Qualifier</a> ] <a href="#type">Type</a> <a href="#var_list">VariableList</a> [ "=" <a href="#init_list">InitalizerList</a> ] ";" ;

<a name="qual">Qualifier</a>     = "const" | ( "public" [ "const" ] ) ;

<a name="var_list">VariableList</a>   = <a href="#variable">Variable</a> { "," <a href="#variable">Variable</a> } ;
<a name="variable">Variable</a>       = <a href="#id">identfier</a> [ <a href="#arr_decl">ArrayDecl</a> ] ;
<a name="arr_decl">ArrayDecl</a>      = "[" <a href="#expression">Expression</a> "]" ;

<a name="init_list">InitalizerList</a> = <a href="#initializer">Initializer</a> { "," <a href="#initializer">Initializer</a> } ;
<a name="initializer">Initializer</a>    = <a href="#expression">Expression</a> | <a href="#arr_init">ArrayInit</a> | <a href="#struct_init">StructInit</a> | <a href="#map_init">MapInit</a> ;
<a name="arr_init">ArrayInit</a>      = "{" { <a href="#initializer">Initializer</a> "," } "}"
<a name="struct_init">StructInit</a>     = "{" { <a href="#initializer">Initializer</a> "," } "}"
<a name="map_init">MapInit</a>        = "{" { "{" <a href="#initializer">Initializer</a> "," <a href="#initializer">Initializer</a> "}" "," } "}"
</pre>

### Expressions

<pre>
Expression = AssignExp ;
ExpressionList = Expression { "," Expression } ;
</pre>

#### Assignement Expressions

<pre>
AssignExp = ExpressionList "=" ExpressionList | UnaryExp AssignOp SqlExp ;
AssignOp  = "+=" | "-=" | "*=" | "/=" | "&=" | "|=" | "^=" | ">>=" | "<<=" ;
</pre>

#### SQL Expressions

<pre>
SqlExp = TernaryExp | Sqls ;
Sqls   = ? insert statement ? |
         ? update statement ? |
         ? delete statement ? |
         ? select statement ? ;
</pre>

#### Ternary Expressions

<pre>
TernaryExp = LogicalExp | 
             LogicalExp "?" SqlExp ":" TernaryExp ;
</pre>

#### Logical Expressions

<pre>
LogicalExp = BitwiseExp | 
             LogicalExp "&&" BitwiseExp |
             LogicalExp "||" BitwiseExp ;
</pre>

#### Bitwise Expressions

<pre>
BitwiseExp = ComparisonExp | 
             BitwsiseExp "|" ComparisonExp |
             BitwsiseExp "^" ComparisonExp |
             BitwsiseExp "&" ComparisonExp ;
</pre>

#### Comparison Expressions

<pre>
ComparisonExp = ShiftExp | 
                ComparisonExp "==" ShiftExp |
                ComparisonExp "!=" ShiftExp |
                ComparisonExp "<" ShiftExp |
                ComparisonExp ">" ShiftExp |
                ComparisonExp "<=" ShiftExp |
                ComparisonExp ">=" ShiftExp ;
</pre>

#### Shift Expressions

<pre>
ShiftExp = ArithmeticExp | 
           ShiftExp ">>" ArithmeticExp |
           ShiftExp "<<" ArithmeticExp ;
</pre>

#### Arithmetic Expressions

<pre>
ArithmeticExp = UnaryExp | 
                ArithmeticExp "+" UnaryExp |
                ArithmeticExp "-" UnaryExp |
                ArithmeticExp "*" UnaryExp |
                ArithmeticExp "/" UnaryExp |
                ArithmeticExp "%" UnaryExp ;
</pre>

#### Unary Expressions

<pre>
UnaryExp = PostExp | 
           "++" UnaryExp |
           "--" UnaryExp |
           "+" UnaryExp |
           "-" UnaryExp |
           "!" UnaryExp ;
</pre>

#### Post Expressions

<pre>
PostExp = PrimaryExp | 
          PostExp "[" TernaryExp "]" |
          PostExp "(" [ ArgumentList ] ")" |
          PostExp "." identifier |
          PostExp "++" |
          PostExp "--" ;
</pre>

#### Primary Expressions

<pre>
PrimarExp = Literals | identifier | "(" Expression ")"
Literals  = bool | integer | float | string | identifier | "null"
</pre>

### Statements

<pre>
Statement = ExpressionStmt | LabelStmt | IfStmt | LoopStmt | SwitchStmt | 
            JumpStmt | DdlStmt | BlockStmt ;
</pre>

#### Expression Statements

<pre>
ExpressionStmt = ";" | Expression ";" ;
</pre>

#### Label Statements

<pre>
LabelStmt = identifier ":" | Statement ;
</pre>

#### If Statements

<pre>
IfStmt       = "if" "(" Expression ")" BlockStmt { ElseIfClause } [ ElseClause ] ;
ElseIfClause = "else" IfStmt ;
ElseClause   = "else" BlockStmt ;
</pre>

#### Loop Statements

<pre>
LoopStmt      = ForLoopStmt | ArrayLoopStmt ;
ForLoopStmt   = "for" BlockStmt |
                "for" "(" LogicalExp ")" BlockStmt |
                "for" "(" InitExp [ ExpressionStmt ] [ Expression ] ")" BlockStmt ;
InitExp       = ExpressionStmt | VariableDecl ;
ArrayLoopStmt = "for" "(" InitExp "in" PostExp ")" BlockStmt ;
</pre>

#### Switch Statements

<pre>
SwitchStmt = "switch" [ "(" Expression ")" ] "{" CaseList "}" ;
CaseList   = "case" ComparisonExp ":" { Statement } |
             "default" ":" { Statement } ;
</pre>

#### Jump Statements

<pre>
JumpStmt = "continue" ";" |
           "break" ";" |
           "return" [ Expression ] ";" |
           "goto" identifier ";" ;
</pre>

#### DDL Statements

<pre>
DdlStmt = ? create index ... ? |
          ? create table ... ? |
          ? drop index ... ? |
          ? drop table ... ? ;
</pre>

#### Block Statements

<pre>
BlockStmt = "{" { VariableDecl | StructDecl | Statement } "}" ;
</pre>

### Constructors

<pre>
ConstructorDecl = identifier "(" [ ParameterList ] ")" BlockStmt ;
ParameterList   = ParameterDecl { "," ParameterDecl } ;
ParameterDecl   = [ "const" ] Type identifier ;
</pre>

### Functions

<pre>
FunctionDecl   = Modifier "func" identifier "(" [ ParameterList ] ")" [ ReturnTypeList ] BlockStmt ;
Modifier       = "public" [ ( "payable" | "readonly" ) ] ;
ReturnTypeList = Type { "," Type } ;
</pre>
