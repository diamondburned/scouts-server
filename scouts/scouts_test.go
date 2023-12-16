package scouts

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	_ "embed"
)

//go:embed scouts_tests.txt
var testsFile string

type ExpectingResult int

const ExpectingTrue ExpectingResult = -1

const (
	ExpectingFalse ExpectingResult = iota
	ExpectingFalseAt1
	ExpectingFalseAt2
	// ...
	// also don't use the AtX constants lol, just use an equality check
)

type TestCase struct {
	Name   string
	Turns  []PastTurn
	Expect ExpectingResult
}

func parseTestCases(testsFile string) ([]TestCase, error) {
	blocks := strings.Split(testsFile, "\n\n")
	testCases := make([]TestCase, 0, len(blocks))
	for i, block := range blocks {
		block = strings.TrimSpace(block)
		testCase, err := parseTestCase(block)
		if err != nil {
			return nil, fmt.Errorf("invalid test case %d: %w", i, err)
		}
		testCases = append(testCases, testCase)
	}
	return testCases, nil
}

func parseTestCase(testBlock string) (TestCase, error) {
	var testCase TestCase
	for _, line := range strings.Split(testBlock, "\n") {
		arg0, argv, ok := strings.Cut(line, ":")
		if !ok {
			return TestCase{}, fmt.Errorf("line must be in the form of \"arg0: argv\": %q", line)
		}

		arg0 = strings.TrimSpace(arg0)
		argv = strings.TrimSpace(argv)

		switch arg0 {
		case "test":
			testCase.Name = argv

		case "expect":
			switch {
			case argv == "true":
				testCase.Expect = ExpectingTrue
			case argv == "false":
				testCase.Expect = ExpectingFalse
			case strings.HasPrefix(argv, "false at "):
				argv = strings.TrimPrefix(argv, "false at ")
				argv = strings.TrimSpace(argv)
				i, err := strconv.Atoi(argv)
				if err != nil {
					return TestCase{}, fmt.Errorf("invalid number for false at statement: %q", argv)
				}
				if i < 1 {
					return TestCase{}, fmt.Errorf("invalid number for false at statement: %q", argv)
				}
				testCase.Expect = ExpectingResult(i)
			default:
				return TestCase{}, fmt.Errorf("invalid expect statement: %q", argv)
			}

		case "player1", "player2", "playerA", "playerB":
			moves, err := ParseMoves(argv)
			if err != nil {
				return TestCase{}, fmt.Errorf("invalid move %q: %w", argv, err)
			}

			var player Player
			switch arg0 {
			case "player1", "playerA":
				player = PlayerA
			case "player2", "playerB":
				player = PlayerB
			}

			testCase.Turns = append(testCase.Turns, PastTurn{
				Player: player,
				Moves:  moves,
			})

		default:
			return TestCase{}, fmt.Errorf("unknown directive: %q", arg0)
		}
	}

	if testCase.Name == "" {
		return TestCase{}, fmt.Errorf("missing test name")
	}

	return testCase, nil
}

func TestGame(t *testing.T) {
	testCases, err := parseTestCases(testsFile)
	if err != nil {
		t.Fatal("failed to parse test cases:", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			g := NewGame()
			for i, turn := range testCase.Turns {
				for _, move := range turn.Moves {
					if err := g.Apply(turn.Player, move); err != nil {
						if testCase.Expect == ExpectingTrue {
							t.Fatalf(
								"failed to apply turn %d, move %q for player %s: %v",
								i+1, move, turn.Player, err)
						}
						if int(testCase.Expect) != i+1 {
							t.Fatalf(
								"failed to apply turn %d, move %q for player %s: %v",
								i+1, move, turn.Player, err)
						}
						return
					}
					t.Logf(
						"applied turn %d, move %q for player %s:\n%s",
						i+1, move, turn.Player, FormatBoard(g.Board()))
				}
			}
		})
	}
}
