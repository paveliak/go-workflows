package main

import (
	"testing"

	"github.com/paveliak/go-workflows/tester"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Workflow(t *testing.T) {
	tester := tester.NewWorkflowTester[string](Workflow1)

	tester.OnActivity(Activity1, mock.Anything, 35, 12).Return(47, nil)

	tester.Execute("Hello world" + uuid.NewString())

	require.True(t, tester.WorkflowFinished())

	wr, werr := tester.WorkflowResult()
	require.Equal(t, "result", wr)
	require.Empty(t, werr)
	tester.AssertExpectations(t)
}
