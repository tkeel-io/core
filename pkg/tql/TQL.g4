 /*
 Copyright © 2021 NAME HERE <EMAIL ADDRESS>

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 */

/*
 * 1. Support select
 */

grammar TQL;

// 1. Tokens & KeyWord
// 1.1 KeyWord
INSERT:                 I N S E R T;
INTO:                   I N T O;
AS:                     A S;
AND:                    A N D;
CASE:                   C A S E;
ELSE:                   E L S E;
END:                    E N D;
EQ:                     E Q     | '=';
FROM:                   F R O M;
GT:                     G T     | '>';
GTE:                    G T E   | '>' '=';
LT:                     L T     | '<';
LTE:                    L T E   | '<' '=';
NE:                     N E     | '!' '=' | '<' '>';
NOT:                    N O T   | '!';
NULL:                   N U L L;
OR:                     O R;
SELECT:                 S E L E C T;
THEN:                   T H E N;
WHERE:                  W H E R E;
WHEN:                   W H E N;
GROUP:                  G R O U P;
BY:                     B Y;
TUMBLINGWINDOW:         T U M B L I N G W I N D O W;
HOPPINGWINDOW:          H O P P I N G W I N D O W;
SLIDINGWINDOW:          S L I D I N G W I N D O W;
SESSIONWINDOW:          S E S S I O N W I N D O W;
DD:                     D D;
HH:                     H H;
MI:                     M I;
SS:                     S S;
MS:                     M S;


// 1.2 Token
MUL:                '*';
DIV:                '/';
MOD:                '%';
ADD:                '+';
SUB:                '-';
DOT:                '.';
TRUE:               T R U E;
FALSE:              F A L S E;
ENTITYNAME:         [a-zA-Z_#*0-9]([a-zA-Z_\-#$@]+[0-9]* | [0-9]*[a-zA-Z_\-#$@]+);
PROPERTYNAME:       '.' [a-zA-Z_#][a-zA-Z_\-#$@0-9.]*;
TARGETENTITY:       [a-zA-Z_#*][a-zA-Z_\-#$@0-9]*;
NUMBER:             '0' | [1-9][0-9]* ;
INTEGER:            ('+' | '-')? NUMBER;
FLOAT:              ('+' | '-')? (NUMBER+ DOT NUMBER+ |  NUMBER+ DOT | DOT NUMBER+);
STRING:             '\'' (~'\'' | '\'\'')* '\'';
WHITESPACE:         [ \r\n\t]+ -> skip;




// 2. Rules
root
    : INSERT INTO targetEntity SELECT fields EOF;


// 2.1 Select
fields
    : expr (',' expr)*
    ;

targetEntity
    : ENTITYNAME
    ;

expr
    : sourceEntity                                  # Expression
    | sourceEntity+ AS targetProperty+              # Expression
    | expr op=('*'|'/'|'%') expr                    # DummyMulDiv
    | expr op=('+'|'-') expr                        # DummyAddSub
    | expr op=(EQ | GT | LT | GTE | LTE | NE) expr  # DummyCompareValue
    ;



// 2.1 entity
sourceEntity
    : '*'
    | ENTITYNAME (PROPERTYNAME)?
    ;

targetProperty
    : ENTITYNAME
    ;


computing
    : numExp EOF;

numExp
   : numExp op=('*'|'/') numExp # MulDiv
   | numExp op=('+'|'-') numExp # AddSub
   | numExp op=(EQ | GT | LT | GTE | LTE | NE) numExp  # CompareValue
   | NUMBER                             # Number
   ;

fragment A: [aA];
fragment B: [bB];
fragment C: [cC];
fragment D: [dD];
fragment E: [eE];
fragment F: [fF];
fragment G: [gG];
fragment H: [hH];
fragment I: [iI];
fragment J: [jJ];
fragment K: [kK];
fragment L: [lL];
fragment M: [mM];
fragment N: [nN];
fragment O: [oO];
fragment P: [pP];
fragment Q: [qQ];
fragment R: [rR];
fragment S: [sS];
fragment T: [tT];
fragment U: [uU];
fragment V: [vV];
fragment W: [wW];
fragment X: [xX];
fragment Y: [yY];
fragment Z: [zZ];