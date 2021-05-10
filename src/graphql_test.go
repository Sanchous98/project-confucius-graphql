package src

import (
	"fmt"
	tools "github.com/bhoriuchi/graphql-go-tools"
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	fakeWeb              = new(GraphQL)
	testInvalidDirective string
	testDirectives       = map[string]interface{}{
		"testSchemaDirective":               testSchemaDirective,
		"testScalarDirective":               testScalarDirective,
		"testObjectDirective":               testObjectDirective,
		"testFieldDefinitionDirective":      testFieldDefinitionDirective,
		"testArgumentDefinitionDirective":   testArgumentDefinitionDirective,
		"testInterfaceDirective":            testInterfaceDirective,
		"testUnionDirective":                testUnionDirective,
		"testEnumDirective":                 testEnumDirective,
		"testEnumValueDirective":            testEnumValueDirective,
		"testInputObjectDirective":          testInputObjectDirective,
		"testInputFieldDefinitionDirective": testInputFieldDefinitionDirective,
	}
)

func testSchemaDirective(*graphql.SchemaConfig, map[string]interface{})                         {}
func testScalarDirective(*graphql.ScalarConfig, map[string]interface{})                         {}
func testObjectDirective(*graphql.ObjectConfig, map[string]interface{})                         {}
func testFieldDefinitionDirective(*graphql.Field, map[string]interface{})                       {}
func testArgumentDefinitionDirective(*graphql.ArgumentConfig, map[string]interface{})           {}
func testInterfaceDirective(*graphql.InterfaceConfig, map[string]interface{})                   {}
func testUnionDirective(*graphql.UnionConfig, map[string]interface{})                           {}
func testEnumDirective(*graphql.EnumConfig, map[string]interface{})                             {}
func testEnumValueDirective(*graphql.EnumValueConfig, map[string]interface{})                   {}
func testInputObjectDirective(*graphql.InputObjectConfig, map[string]interface{})               {}
func testInputFieldDefinitionDirective(*graphql.InputObjectFieldConfig, map[string]interface{}) {}
func testInvalidFuncDirective()                                                                 {}

func TestAddResolvers(t *testing.T) {
	fakeWeb.directives = make(tools.SchemaDirectiveVisitorMap)

	for name, directive := range testDirectives {
		fakeWeb.AddDirective(name, directive)
	}

	assert.True(t, fakeWeb.DirectiveExists("testSchemaDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testScalarDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testObjectDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testFieldDefinitionDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testArgumentDefinitionDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testInterfaceDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testUnionDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testEnumDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testEnumValueDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testInputObjectDirective"))
	assert.True(t, fakeWeb.DirectiveExists("testInputFieldDefinitionDirective"))
	assert.Panics(t, func() {
		fakeWeb.AddDirective("testInvalidDirective", testInvalidDirective)
	})
	assert.Panics(t, func() {
		fakeWeb.AddDirective("testInvalidFuncDirective", testInvalidFuncDirective)
	})
}

func TestDropDirective(t *testing.T) {
	TestAddResolvers(t)
	fakeWeb.DropDirective("testSchemaDirective")
	assert.False(t, fakeWeb.DirectiveExists("testSchemaDirective"))
}

func TestResolveDirective(t *testing.T) {
	fakeSchema := []byte("# @IsGranted limits access to users, using roles\ndirective @isGranted(roles: [Roles]!) on QUERY | MUTATION | FIELD_DEFINITION\n\nenum Roles {\n    ANONYMOUS,\n    STUDENT,\n    TEACHER,\n    ROOT\n}\n\nenum OrganisationType {\n    SCHOOL,\n    UNIVERSITY,\n    COLLEGE\n}\n\ntype User {\n    id: ID!\n    email: String!\n    firstName: String!\n    lastName: String!\n    role: Roles\n    organisation: Organisation!\n    group: Group!\n    courses: [Course]\n}\n\ntype Group {\n    id: ID!\n    name: String!\n    headman: User!\n    students: [User]!\n}\n\ntype Organisation {\n    id: ID!\n    type: OrganisationType!\n    name: String!\n    members: [User]!\n}\n\ntype Course {\n    id: ID!\n    name: String!\n    organisation: Organisation!\n    members: [User]\n}\n\ntype Query {\n    user(id: ID!): User @isGranted(roles: [STUDENT, TEACHER, ROOT])\n    course(id: ID!): Course\n}")
	g := new(GraphQL)
	g.directives = make(tools.SchemaDirectiveVisitorMap)

	schema := g.resolveSchema(fakeSchema)
	fmt.Println(schema.Directives()[3])
}
