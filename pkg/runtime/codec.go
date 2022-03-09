package runtime

type PatchOp string

const (
	OpUndef   PatchOp = "undefine"
	OpAdd     PatchOp = "add"
	OpTest    PatchOp = "test"
	OpCopy    PatchOp = "copy"
	OpMove    PatchOp = "move"
	OpMerge   PatchOp = "merge"
	OpRemove  PatchOp = "remove"
	OpReplace PatchOp = "replace"
)
