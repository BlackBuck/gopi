package main

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/BlackBuck/pcom-go"
)

var jsonValue parser.Parser

func whitespace() parser.Parser {
	return parser.Many0(
		parser.Or(
			parser.Or(
				parser.CharParser(' '),
				parser.CharParser('\n'),
			),
			parser.Or(
				parser.CharParser('\t'),
				parser.CharParser('\r'),
			),
		),
	)
}

func jsonString() parser.Parser {
	stringChar := func() parser.Parser {
		return func(curState parser.State) (parser.Result, error) {
			c, err := curState.PeekChar()
			if err != nil {
				return parser.NewResult(nil, curState), fmt.Errorf("unexpected end of input")
			}

			if c != '"' && c != '\\' {
				return parser.NewResult(
					string(c),
					curState.Advance(1),
				), nil
			}

			return parser.NewResult(nil, curState), fmt.Errorf("unexpected character %s", string(c))
		}
	}

	return parser.Map(
		parser.Between(
			parser.CharParser('"'),
			parser.Many0(stringChar()),
			parser.CharParser('"'),
		),
		func(chars interface{}) string {
			var sb strings.Builder
			for _, c := range chars.([]interface{}) {
				sb.WriteString(c.(string))
			}
			return sb.String()
		},
	)
}

// parse a number
func jsonNumber() parser.Parser {
	digit := func(curState parser.State) (parser.Result, error) {

		ch, err := curState.PeekChar()
		if err != nil {
			return parser.NewResult(nil, curState), fmt.Errorf("unexpected end of input")
		}

		if ch >= '0' && ch <= '9' {
			return parser.NewResult(
				string(ch),
				curState.Advance(1),
			), nil
		}

		return parser.NewResult(nil, curState), fmt.Errorf("expected digit")
	}

	return parser.Map(
		parser.Many1(digit),
		func(digits interface{}) float64 {
			var sb strings.Builder
			for _, d := range digits.([]interface{}) {
				sb.WriteString(d.(string))
			}

			num, _ := strconv.ParseFloat(sb.String(), 64)

			return num
		},
	)

}

// parse a json array
func jsonArray() parser.Parser {
	return parser.Map(
		parser.Between(
			parser.Seq(parser.CharParser('['), whitespace()),
			parser.Many0(
				parser.Seq(
					parser.Lazy(func() parser.Parser { return jsonValue }),
					parser.Many0(
						parser.Seq(
							parser.CharParser(','),
							whitespace(),
							parser.Lazy(func() parser.Parser { return jsonValue }),
						),
					),
				),
			),
			parser.Seq(whitespace(), parser.CharParser(']')),
		),

		func(val interface{}) []interface{} {
			if len(val.([]interface{})) == 0 {
				return []interface{}{}
			}
			result := make([]interface{}, 0)
			seqResults := val.([]interface{})
			result = append(result, seqResults[0].([]interface{})[0])
			
			// items of the second seq in the form [[',', ' ', jsonvalue], ...]
			restItems := seqResults[0].([]interface{})[1].([]interface{})
			for _, item := range restItems {
				itemSeq := item.([]interface{})
				result = append(result, itemSeq[2])
			}

			return result
		},
	)
}

func ParseJSON(input string) (interface{}, error) {
	jsonValue = parser.Or(
		parser.Or(
			jsonString(),
			jsonNumber(),
		),
		parser.Lazy(jsonArray),
	)

	res, err := jsonValue(parser.NewState(input, 0))

	fmt.Println(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}


func main() {
	// Test cases
	inputs := []string{
		`123`,
		`"Hello World"`,
		`[ 1, 2, "Hello World" ]`,
		`[ 1, 2, [ 1, 3 ] ]`,
		`[ ]`,
	}

	fmt.Printf("Test cases: \n %s", inputs)
	for _, input := range inputs {
		result, err := ParseJSON(input)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", input, err)
			continue
		}
		fmt.Printf("Successfully parsed %s: %v\n", input, result)
	}
}