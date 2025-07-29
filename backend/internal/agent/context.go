package agent

// define prompt stuff. Truncate. builder. token counter.

// define SetContext() GetContext() and higher-level UpdateContext() enforce type rules on the keys used. use:
// shaikh:user:{user_id}:session:{session_id}:context
// Define what type of value goes in for context. Add created_at, updated_at, context which points to a parseable slice of structs (ayat, documents, previous sessions, memories, question & answer), and allocate token limits for the parts.
