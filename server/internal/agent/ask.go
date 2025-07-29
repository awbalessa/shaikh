package agent

import (
	"context"
	"iter"
)

// app.Agent.Ask(ctx context.Context, prompt string) iter.Seq2
// ...
// call searcher. If tool call gotten with name, call searcher.callFn(fn fname), which will receive func

func (a *Agent) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error]
