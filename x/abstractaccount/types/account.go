package types

import "fmt"

func AccountLabel(sender string, codeID uint64) string {
	// Ideally we have a unique label for each abstract account. The current one
	// obvious isn't unique.
	//
	// An option is `abstractaccount/blockHeight/txIndex/msgIndex` but I haven't
	// figured how to determine the txIndex and msgIndex yet.
	return fmt.Sprintf("abstractaccount/%s/%d", sender, codeID)
}
