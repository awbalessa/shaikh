package app

func (a *App) Run() error {
	// Embed query -> Retrieve docs -> Prompt Gemini -> Print result
	query := "بني إسرائيل خالفوا أوامر الله وتعلموا السحر في عهد سليمان"

	embedRes, err := a.EmbedQuery(query)
	if err != nil {
		return err
	}

	topDocs, err := a.RetrieveRelevantDocs(embedRes)
	if err != nil {
		return err
	}

	result, err := a.GenerateAnswer(query, topDocs)
	if err != nil {
		return err
	}

	a.Logger.Println("Gemini Answer:\n", result)
	return nil
}
