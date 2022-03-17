package yqlib

import (
	"fmt"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

type compareTypePref struct {
	OrEqual bool
	Greater bool
}

func compareOperator(d *dataTreeNavigator, context Context, expressionNode *ExpressionNode) (Context, error) {
	log.Debugf("-- compareOperator")
	prefs := expressionNode.Operation.Preferences.(compareTypePref)
	return crossFunction(d, context.ReadOnlyClone(), expressionNode, compare(prefs), true)
}

func compare(prefs compareTypePref) func(d *dataTreeNavigator, context Context, lhs *CandidateNode, rhs *CandidateNode) (*CandidateNode, error) {
	return func(d *dataTreeNavigator, context Context, lhs *CandidateNode, rhs *CandidateNode) (*CandidateNode, error) {
		log.Debugf("-- compare cross function")
		if lhs == nil && rhs == nil {
			owner := &CandidateNode{}
			return createBooleanCandidate(owner, prefs.OrEqual), nil
		} else if lhs == nil {
			log.Debugf("lhs nil, but rhs is not")
			return createBooleanCandidate(rhs, false), nil
		} else if rhs == nil {
			log.Debugf("rhs nil, but rhs is not")
			return createBooleanCandidate(lhs, false), nil
		}

		lhs.Node = unwrapDoc(lhs.Node)
		rhs.Node = unwrapDoc(rhs.Node)

		switch lhs.Node.Kind {
		case yaml.MappingNode:
			return nil, fmt.Errorf("maps not yet supported for comparison")
		case yaml.SequenceNode:
			return nil, fmt.Errorf("arrays not yet supported for comparison")
		default:
			if rhs.Node.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("%v (%v) cannot be subtracted from %v", rhs.Node.Tag, rhs.Path, lhs.Node.Tag)
			}
			target := lhs.CreateReplacement(&yaml.Node{})
			boolV, err := compareScalars(context, prefs, lhs.Node, rhs.Node)

			return createBooleanCandidate(target, boolV), err
		}
	}
}

func compareScalars(context Context, prefs compareTypePref, lhs *yaml.Node, rhs *yaml.Node) (bool, error) {
	lhsTag := guessTagFromCustomType(lhs)
	rhsTag := guessTagFromCustomType(rhs)

	// isDateTime := lhs.Tag == "!!timestamp"
	// // if the lhs is a string, it might be a timestamp in a custom format.
	// if lhsTag == "!!str" && context.GetDateTimeLayout() != time.RFC3339 {
	// 	_, err := time.Parse(context.GetDateTimeLayout(), lhs.Value)
	// 	isDateTime = err == nil
	// }

	if lhsTag == "!!int" && rhsTag == "!!int" {
		_, lhsNum, err := parseInt(lhs.Value)
		if err != nil {
			return false, err
		}
		_, rhsNum, err := parseInt(rhs.Value)
		if err != nil {
			return false, err
		}

		if prefs.OrEqual && lhsNum == rhsNum {
			return true, nil
		}
		if prefs.Greater {
			return lhsNum > rhsNum, nil
		}
		return lhsNum < rhsNum, nil
	} else if (lhsTag == "!!int" || lhsTag == "!!float") && (rhsTag == "!!int" || rhsTag == "!!float") {
		lhsNum, err := strconv.ParseFloat(lhs.Value, 64)
		if err != nil {
			return false, err
		}
		rhsNum, err := strconv.ParseFloat(rhs.Value, 64)
		if err != nil {
			return false, err
		}
		if prefs.OrEqual && lhsNum == rhsNum {
			return true, nil
		}
		if prefs.Greater {
			return lhsNum > rhsNum, nil
		}
		return lhsNum < rhsNum, nil
	}
	return false, fmt.Errorf("not yet supported")
}
