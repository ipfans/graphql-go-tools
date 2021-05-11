package testsgo

import (
	"testing"
)

func TestUniqueVariableNamesRule(t *testing.T) {

	expectErrors := func(queryStr string) ResultCompare {
		return ExpectValidationErrors("UniqueVariableNamesRule", queryStr)
	}

	expectValid := func(queryStr string) {
		expectErrors(queryStr)([]Err{})
	}

	t.Run("Validate: Unique variable names", func(t *testing.T) {
		t.Run("unique variable names", func(t *testing.T) {
			expectValid(`
      query A($x: Int, $y: String) { __typename }
      query B($x: String, $y: Int) { __typename }
    `)
		})

		t.Run("duplicate variable names", func(t *testing.T) {
			expectErrors(`
      query A($x: Int, $x: Int, $x: String) { __typename }
      query B($x: String, $x: Int) { __typename }
      query C($x: Int, $x: Int) { __typename }
    `)([]Err{
				{
					message: `There can be only one variable named "$x".`,
					locations: []Loc{
						{line: 2, column: 16},
						{line: 2, column: 25},
					},
				},
				{
					message: `There can be only one variable named "$x".`,
					locations: []Loc{
						{line: 2, column: 16},
						{line: 2, column: 34},
					},
				},
				{
					message: `There can be only one variable named "$x".`,
					locations: []Loc{
						{line: 3, column: 16},
						{line: 3, column: 28},
					},
				},
				{
					message: `There can be only one variable named "$x".`,
					locations: []Loc{
						{line: 4, column: 16},
						{line: 4, column: 25},
					},
				},
			})
		})
	})

}
