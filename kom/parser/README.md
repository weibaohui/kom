docker run --rm \
-v `pwd`:/work antlr/antlr4 \
-o parser -Dlanguage=Go SQLiteLexer.g4 SQLiteParser.g4