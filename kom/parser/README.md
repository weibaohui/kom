git clone https://github.com/antlr/antlr4.git
cd antlr4/docker
docker build -t antlr/antlr4 --platform linux/amd64 .

docker run --rm \
-v `pwd`:/work antlr/antlr4 \
-o parser -Dlanguage=Go SQLiteLexer.g4 SQLiteParser.g4


# #
#IDENTIFIER 
#: ( [A-Za-z_] [A-Za-z_0-9]* ) ( '.' [A-Za-z_] [A-Za-z_0-9]* )* ( '[' ~']'* ']' )*
#;

[//]: # (以下示例均会被正确识别为 IDENTIFIER：)

[//]: # ()
[//]: # (status)

[//]: # (status.addresses)

[//]: # (status.addresses[type=InternalIP])

[//]: # (status.addresses[type=InternalIP].address)

[//]: # (metadata.annotations['key.name'])