package bramble

import (
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

var IdFieldName = "id"
var IdFieldType = "ID!"

const (
	nodeRootFieldName      = "node"
	nodeInterfaceName      = "Node"
	serviceObjectName      = "Service"
	serviceRootFieldName   = "service"
	boundaryDirectiveName  = "boundary"
	namespaceDirectiveName = "namespace"
	skipMergeDirectiveName = "skipMerge"

	queryObjectName        = "Query"
	mutationObjectName     = "Mutation"
	subscriptionObjectName = "Subscription"

	internalServiceName = "__bramble"
)

func isGraphQLBuiltinName(s string) bool {
	return strings.HasPrefix(s, "__")
}

func isIDType(t *ast.Type) bool {
	return isNonNullableTypeNamed(t, "ID")
}

func isNonNullableTypeNamed(t *ast.Type, typename string) bool {
	return t.Name() == typename && t.NonNull
}

func isNullableTypeNamed(t *ast.Type, typename string) bool {
	return t.Name() == typename && !t.NonNull
}
