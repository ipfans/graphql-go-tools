package testsgo

import (
	"testing"
)

func TestKnownArgumentNamesRule(t *testing.T) {

	expectErrors := func(queryStr string) ResultCompare {
		return ExpectValidationErrors("KnownArgumentNamesRule", queryStr)
	}

	expectValid := func(queryStr string) {
		expectErrors(queryStr)([]Err{})
	}

	expectSDLErrors := func(sdlStr string, sch ...string) ResultCompare {
		schema := ""
		if len(sch) > 0 {
			schema = sch[0]
		}
		return ExpectSDLValidationErrors(
			schema,
			"KnownArgumentNamesOnDirectivesRule",
			sdlStr,
		)
	}

	expectValidSDL := func(sdlStr string, schema ...string) {
		expectSDLErrors(sdlStr)([]Err{})
	}

	t.Run("Validate: Known argument names", func(t *testing.T) {
		t.Run("single arg is known", func(t *testing.T) {
			expectValid(`
      fragment argOnRequiredArg on Dog {
        doesKnowCommand(dogCommand: SIT)
      }
    `)
		})

		t.Run("multiple args are known", func(t *testing.T) {
			expectValid(`
      fragment multipleArgs on ComplicatedArgs {
        multipleReqs(req1: 1, req2: 2)
      }
    `)
		})

		t.Run("ignores args of unknown fields", func(t *testing.T) {
			expectValid(`
      fragment argOnUnknownField on Dog {
        unknownField(unknownArg: SIT)
      }
    `)
		})

		t.Run("multiple args in reverse order are known", func(t *testing.T) {
			expectValid(`
      fragment multipleArgsReverseOrder on ComplicatedArgs {
        multipleReqs(req2: 2, req1: 1)
      }
    `)
		})

		t.Run("no args on optional arg", func(t *testing.T) {
			expectValid(`
      fragment noArgOnOptionalArg on Dog {
        isHouseTrained
      }
    `)
		})

		t.Run("args are known deeply", func(t *testing.T) {
			expectValid(`
      {
        dog {
          doesKnowCommand(dogCommand: SIT)
        }
        human {
          pet {
            ... on Dog {
              doesKnowCommand(dogCommand: SIT)
            }
          }
        }
      }
    `)
		})

		t.Run("directive args are known", func(t *testing.T) {
			expectValid(`
      {
        dog @skip(if: true)
      }
    `)
		})

		t.Run("field args are invalid", func(t *testing.T) {
			expectErrors(`
      {
        dog @skip(unless: true)
      }
    `)([]Err{
				{
					message:   `Unknown argument "unless" on directive "@skip".`,
					locations: []Loc{{line: 3, column: 19}},
				},
			})
		})

		t.Run("directive without args is valid", func(t *testing.T) {
			expectValid(`
      {
        dog @onField
      }
    `)
		})

		t.Run("arg passed to directive without arg is reported", func(t *testing.T) {
			expectErrors(`
      {
        dog @onField(if: true)
      }
    `)([]Err{
				{
					message:   `Unknown argument "if" on directive "@onField".`,
					locations: []Loc{{line: 3, column: 22}},
				},
			})
		})

		t.Run("misspelled directive args are reported", func(t *testing.T) {
			expectErrors(`
      {
        dog @skip(iff: true)
      }
    `)([]Err{
				{
					message:   `Unknown argument "iff" on directive "@skip". Did you mean "if"?`,
					locations: []Loc{{line: 3, column: 19}},
				},
			})
		})

		t.Run("invalid arg name", func(t *testing.T) {
			expectErrors(`
      fragment invalidArgName on Dog {
        doesKnowCommand(unknown: true)
      }
    `)([]Err{
				{
					message:   `Unknown argument "unknown" on field "Dog.doesKnowCommand".`,
					locations: []Loc{{line: 3, column: 25}},
				},
			})
		})

		t.Run("misspelled arg name is reported", func(t *testing.T) {
			expectErrors(`
      fragment invalidArgName on Dog {
        doesKnowCommand(DogCommand: true)
      }
    `)([]Err{
				{
					message:   `Unknown argument "DogCommand" on field "Dog.doesKnowCommand". Did you mean "dogCommand"?`,
					locations: []Loc{{line: 3, column: 25}},
				},
			})
		})

		t.Run("unknown args amongst known args", func(t *testing.T) {
			expectErrors(`
      fragment oneGoodArgOneInvalidArg on Dog {
        doesKnowCommand(whoKnows: 1, dogCommand: SIT, unknown: true)
      }
    `)([]Err{
				{
					message:   `Unknown argument "whoKnows" on field "Dog.doesKnowCommand".`,
					locations: []Loc{{line: 3, column: 25}},
				},
				{
					message:   `Unknown argument "unknown" on field "Dog.doesKnowCommand".`,
					locations: []Loc{{line: 3, column: 55}},
				},
			})
		})

		t.Run("unknown args deeply", func(t *testing.T) {
			expectErrors(`
      {
        dog {
          doesKnowCommand(unknown: true)
        }
        human {
          pet {
            ... on Dog {
              doesKnowCommand(unknown: true)
            }
          }
        }
      }
    `)([]Err{
				{
					message:   `Unknown argument "unknown" on field "Dog.doesKnowCommand".`,
					locations: []Loc{{line: 4, column: 27}},
				},
				{
					message:   `Unknown argument "unknown" on field "Dog.doesKnowCommand".`,
					locations: []Loc{{line: 9, column: 31}},
				},
			})
		})

		t.Run("within SDL", func(t *testing.T) {
			t.Run("known arg on directive defined inside SDL", func(t *testing.T) {
				expectValidSDL(`
        type Query {
          foo: String @test(arg: "")
        }

        directive @test(arg: String) on FIELD_DEFINITION
      `)
			})

			t.Run("unknown arg on directive defined inside SDL", func(t *testing.T) {
				expectSDLErrors(`
        type Query {
          foo: String @test(unknown: "")
        }

        directive @test(arg: String) on FIELD_DEFINITION
      `)([]Err{
					{
						message:   `Unknown argument "unknown" on directive "@test".`,
						locations: []Loc{{line: 3, column: 29}},
					},
				})
			})

			t.Run("misspelled arg name is reported on directive defined inside SDL", func(t *testing.T) {
				expectSDLErrors(`
        type Query {
          foo: String @test(agr: "")
        }

        directive @test(arg: String) on FIELD_DEFINITION
      `)([]Err{
					{
						message:   `Unknown argument "agr" on directive "@test". Did you mean "arg"?`,
						locations: []Loc{{line: 3, column: 29}},
					},
				})
			})

			t.Run("unknown arg on standard directive", func(t *testing.T) {
				expectSDLErrors(`
        type Query {
          foo: String @deprecated(unknown: "")
        }
      `)([]Err{
					{
						message:   `Unknown argument "unknown" on directive "@deprecated".`,
						locations: []Loc{{line: 3, column: 35}},
					},
				})
			})

			t.Run("unknown arg on overridden standard directive", func(t *testing.T) {
				expectSDLErrors(`
        type Query {
          foo: String @deprecated(reason: "")
        }
        directive @deprecated(arg: String) on FIELD
      `)([]Err{
					{
						message:   `Unknown argument "reason" on directive "@deprecated".`,
						locations: []Loc{{line: 3, column: 35}},
					},
				})
			})

			t.Run("unknown arg on directive defined in schema extension", func(t *testing.T) {
				schema := BuildSchema(`
        type Query {
          foo: String
        }
      `)
				expectSDLErrors(
					`
          directive @test(arg: String) on OBJECT

          extend type Query  @test(unknown: "")
        `,
					schema,
				)([]Err{
					{
						message:   `Unknown argument "unknown" on directive "@test".`,
						locations: []Loc{{line: 4, column: 36}},
					},
				})
			})

			t.Run("unknown arg on directive used in schema extension", func(t *testing.T) {
				schema := BuildSchema(`
        directive @test(arg: String) on OBJECT

        type Query {
          foo: String
        }
      `)
				expectSDLErrors(
					`
          extend type Query @test(unknown: "")
        `,
					schema,
				)([]Err{
					{
						message:   `Unknown argument "unknown" on directive "@test".`,
						locations: []Loc{{line: 2, column: 35}},
					},
				})
			})
		})
	})

}
