OLD_MODULE_NAME="github.com/0xPolygon/polygon-edge"
NEW_MODULE_NAME="github.com/Gabulhas/polygon-external-consensus"
go mod edit -module ${NEW_MODULE_NAME}
find . -type f -name '*.go' -exec sed -i -e 's,{OLD_MODULE_NAME},{NEW_MODULE_NAME},g' {} 
