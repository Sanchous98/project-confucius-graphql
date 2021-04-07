package src

import (
	"encoding/json"
	"fmt"
	"github.com/Sanchous98/project-confucius-base/stdlib"
	"github.com/Sanchous98/project-confucius-base/utils"
	appGraphql "github.com/Sanchous98/project-confucius-graphql/src/graphql"
	tools "github.com/bhoriuchi/graphql-go-tools"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
)

var preDefinedDirectives = map[string]interface{}{
	"isGranted": appGraphql.IsGranted,
}

const graphQLConfigPath = "config/graphql.yaml"
const graphQLEntryPointGroup = "graphql"

type (
	// Type aliases for visitor functions
	VisitSchema               = func(*graphql.SchemaConfig, map[string]interface{})
	VisitScalar               = func(*graphql.ScalarConfig, map[string]interface{})
	VisitObject               = func(*graphql.ObjectConfig, map[string]interface{})
	VisitFieldDefinition      = func(*graphql.Field, map[string]interface{})
	VisitArgumentDefinition   = func(*graphql.ArgumentConfig, map[string]interface{})
	VisitInterface            = func(*graphql.InterfaceConfig, map[string]interface{})
	VisitUnion                = func(*graphql.UnionConfig, map[string]interface{})
	VisitEnum                 = func(*graphql.EnumConfig, map[string]interface{})
	VisitEnumValue            = func(*graphql.EnumValueConfig, map[string]interface{})
	VisitInputObject          = func(*graphql.InputObjectConfig, map[string]interface{})
	VisitInputFieldDefinition = func(*graphql.InputObjectFieldConfig, map[string]interface{})

	graphQLConfig struct {
		SchemaPath string `yaml:"schema_path"`
	}

	GraphEntryPoint struct {
		stdlib.EntryPoint
		SchemaPath string
	}

	GraphQL struct {
		config      *graphQLConfig
		directives  tools.SchemaDirectiveVisitorMap
		Web         *stdlib.Web `inject:""`
		Log         *stdlib.Log `inject:""`
		entryPoints []*GraphEntryPoint
	}
)

func (c *graphQLConfig) Unmarshall() error {
	return utils.Unmarshall(c, graphQLConfigPath, yaml.Unmarshal)
}

func (g *GraphQL) Constructor() {
	g.config = new(graphQLConfig)
	err := g.config.Unmarshall()

	if err != nil {
		panic(err)
	}

	if g.directives == nil {
		g.directives = make(tools.SchemaDirectiveVisitorMap)
	}
	// Add pre-defined directives
	for name, directive := range preDefinedDirectives {
		g.AddDirective(name, directive)
	}

	g.AddEntryPoint(&GraphEntryPoint{
		EntryPoint: stdlib.EntryPoint{
			Routes: []*stdlib.Route{
				{stdlib.MethodGet, "/api", g.queryHandler},
				{stdlib.MethodGet, "/graphiql", g.HandleGraphiQL(g.config.SchemaPath)},
				{stdlib.MethodPost, "/graphiql", g.HandleGraphiQL(g.config.SchemaPath)},
			},
			Name: "main",
		},
		SchemaPath: g.config.SchemaPath,
	})
}

func (g *GraphQL) Destructor() {}

// resolveSchema initializes GraphQL server configuration, based on schema file and predefined resolvers and directives
func (g *GraphQL) resolveSchema(schemaContent []byte) *graphql.Schema {
	schema, err := tools.MakeExecutableSchema(tools.ExecutableSchema{
		TypeDefs:         string(schemaContent),
		SchemaDirectives: g.directives,
	})

	if err != nil {
		g.Log.Alert(fmt.Errorf("failed to parse schema, error: %v", err))
	}

	return &schema
}

// HandleGraphiQL provides GraphiQL playground
func (g *GraphQL) HandleGraphiQL(schemaPath string) fasthttp.RequestHandler {
	file, err := filepath.Abs(schemaPath)

	if err != nil {
		g.Log.Alert(err)
	}

	schema, err := ioutil.ReadFile(file)

	if err != nil {
		g.Log.Alert(fmt.Errorf("cannot open schema file: %v\n", err))
	}

	return fasthttpadaptor.NewFastHTTPHandler(handler.New(&handler.Config{
		Schema:     g.resolveSchema(schema),
		Pretty:     true,
		GraphiQL:   true,
		Playground: true,
	}))
}

// queryHandler is a function that handles GraphQL queries
func (g *GraphQL) queryHandler(context *fasthttp.RequestCtx) {
	b, err := ioutil.ReadFile(g.config.SchemaPath)

	if err != nil {
		g.Log.Alert(fmt.Errorf("schema file doesn't exist"))
	}

	query := context.Request.Body()

	response := graphql.Do(graphql.Params{
		Schema:        *g.resolveSchema(b),
		RequestString: string(query),
	})

	if len(response.Errors) > 0 {
		g.Log.Info(fmt.Errorf("failed to execute graphql operation, errors: %+v", response.Errors))
	}

	//writer.Header().Set("X-CSRF-Token", csrf.Token(context))
	jsonResponse, _ := json.Marshal(response)
	log.Print(context.Write(jsonResponse))
}

// AddDirective adds directive dynamically
func (g *GraphQL) AddDirective(name string, directive interface{}) {
	funcType := reflect.ValueOf(directive)

	if funcType.Kind() != reflect.Func {
		g.Log.Alert(fmt.Errorf("directive should be type of function"))
	}

	newDirective := new(tools.SchemaDirectiveVisitor)

	switch directive := directive.(type) {
	case VisitSchema:
		newDirective.VisitSchema = directive
	case VisitScalar:
		newDirective.VisitScalar = directive
	case VisitObject:
		newDirective.VisitObject = directive
	case VisitFieldDefinition:
		newDirective.VisitFieldDefinition = directive
	case VisitArgumentDefinition:
		newDirective.VisitArgumentDefinition = directive
	case VisitInterface:
		newDirective.VisitInterface = directive
	case VisitUnion:
		newDirective.VisitUnion = directive
	case VisitEnum:
		newDirective.VisitEnum = directive
	case VisitEnumValue:
		newDirective.VisitEnumValue = directive
	case VisitInputObject:
		newDirective.VisitInputObject = directive
	case VisitInputFieldDefinition:
		newDirective.VisitInputFieldDefinition = directive
	default:
		g.Log.Alert(fmt.Errorf("invalid directive definition"))
	}

	g.directives[name] = newDirective
}

// DropDirective drops directive dynamically
func (g *GraphQL) DropDirective(name string) {
	if !g.DirectiveExists(name) {
		return
	}

	delete(g.directives, name)
}

// DirectiveExists checks, if directive exists
func (g *GraphQL) DirectiveExists(name string) bool {
	for directiveName := range g.directives {
		if name == directiveName {
			return true
		}
	}

	return false
}

// AddEntryPoint adds a GraphQL entrypoint dynamically
func (g *GraphQL) AddEntryPoint(entryPoint *GraphEntryPoint) {
	g.entryPoints = append(g.entryPoints, entryPoint)
	g.Web.AddEntryPoint(&entryPoint.EntryPoint, graphQLEntryPointGroup)
}

// DropEntryPoint drops GraphQL entrypoint dynamically
func (g *GraphQL) DropEntryPoint(name string) {
	for index, entryPoint := range g.entryPoints {
		if entryPoint.Name == name {
			g.entryPoints = append(g.entryPoints[:index], g.entryPoints[index+1:]...)
			return
		}
	}
}

// EntryPointExists checks, if GraphQL entrypoint dynamically
func (g *GraphQL) EntryPointExists(name string) bool {
	for _, entryPoint := range g.entryPoints {
		if entryPoint.Name == name {
			return true
		}
	}

	return false
}
