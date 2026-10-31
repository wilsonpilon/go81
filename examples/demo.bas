# Go81 — programa de demonstração (BASIC do ZX81 em texto)
# go81 basic make examples/demo.bas -run 10   ou   go81 shell -c "load examples/demo.bas; typeln R"
10 REM GO81 DEMO
20 PRINT AT 1,13;"%G%O%8%1"
30 PRINT AT 3,6;"EMULADOR ZX81 EM GO"
40 FOR X=0 TO 63
50 PLOT X,25+INT (7*SIN (X/5))
60 NEXT X
70 PRINT AT 16,0;"\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:\:'\.:"
80 PRINT AT 18,3;"%Z%X%8%1 % %T%S%1%0%0%0 % %L%A%M%B%D%A"
90 PRINT AT 20,6;"PRESSIONE UMA TECLA"
100 IF INKEY$="" THEN GOTO 100
110 CLS
120 PRINT "OBRIGADO"
