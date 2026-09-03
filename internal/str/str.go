package str

func SplitStringInChunks(str string, chunkSize int) []string {
	runes := []rune(str)
	runesLength := len(runes)

	if runesLength == 0 {
		return []string{}
	}

	var chunks []string
	for i := 0; i < runesLength; i += chunkSize {
		end := min(i+chunkSize, runesLength)

		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}
