package app

func (a *App) RunRAG(query string) error {
	// Embed query -> Retrieve docs -> Prompt Gemini -> Print result
	vec, err := a.EmbedQuery(query)
	if err != nil {
		return err
	}

	docs, err := a.RetrieveDocuments(vec, 5)
	if err != nil {
		return err
	}

	// // result, err := a.GenerateResponseStream(query, docs)
	// // if err != nil {
	// // 	return err
	// // }

	// a.Logger.Println("Gemini Answer:\n", result)
	// return nil
}
