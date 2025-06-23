package prompt

import (
	"github.com/awbalessa/shaikh/apps/server/internal/database"
)

type Builder struct {
	instructions      []string
	query             string
	documents         []database.CosineSimilarityRow
	includeMetadata   bool
	includeSource     bool
	includeSimilarity bool
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) WithUserQuery(q string) *Builder {
	b.query = q
	return b
}

func (b *Builder) WithDocuments(docs []database.CosineSimilarityRow) *Builder {
	b.documents = docs
	return b
}

func (b *Builder) WithInstructions(instr []string) *Builder {
	b.instructions = instr
	return b
}

func (b *Builder) WithMetadata(include bool) *Builder {
	b.includeMetadata = include
	return b
}

func (b *Builder) WithSource(include bool) *Builder {
	b.includeSource = include
	return b
}

func (b *Builder) WithSimilarity(include bool) *Builder {
	b.includeSimilarity = include
	return b
}

// func (b *Builder) Build() (systemInstructions string, prompt string, err error) {
// 	if b.query == "" {
// 		return "", "", fmt.Errorf("No user query given")
// 	}
// 	if len(b.documents) == 0 {
// 		return "", "", fmt.Errorf("No documents given")
// 	}
// 	if len(b.instructions) == 0 {
// 		return "", "", fmt.Errorf("No system instructions given")
// 	}
// 	var instrBuilder strings.Builder
// 	instrBuilder.WriteString("You are a helpful assistant answering a user's question based only on retrieved documents provided by a retrieval-augmented generation (RAG) system. Follow the instructions below strictly:\n\n")
// 	for _, instr := range b.instructions {
// 		instrBuilder.WriteString(fmt.Sprintf("- %s\n", instr))
// 	}
// 	systemInstructions = instrBuilder.String()

// 	var sb strings.Builder
// 	// Construct prompt
// 	sb.WriteString("--- USER PROMPT ---\n")
// 	sb.WriteString(fmt.Sprintf("%s\n\n", b.query))
// 	// Construct document label
// 	sb.WriteString("--- RETRIEVED DOCUMENTS ---\n")
// 	// Construct documents
// 	for i, doc := range b.documents {
// 		sb.WriteString(fmt.Sprintf("%d. ", i+1))
// 		if b.includeMetadata && len(doc.Metadata) != 0 {
// 			meta, err := models.ExtractMetadata(doc.Metadata)
// 			if err != nil {
// 				return "", "", fmt.Errorf("Error extracting metadata: %v", err)
// 			}
// 			sb.WriteString(fmt.Sprintf("Metadata: %s\n", meta.Describe()))
// 		}
// 		sb.WriteString(fmt.Sprintf("Content: %s\n", doc.RawContent))
// 		if b.includeSimilarity {
// 			sb.WriteString(fmt.Sprintf("\t[similarity: %.2f]\n", doc.Similarity))
// 		}
// 		if b.includeSource {
// 			sb.WriteString(fmt.Sprintf("\t[source: %s]\n", doc.LiteratureSource))
// 		}
// 		sb.WriteString("\n---\n")
// 	}
// 	prompt = sb.String()
// 	err = nil
// 	return
// }
