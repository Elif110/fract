package obj

// Code block instance.
type Block struct {
	Try   func()
	Catch func(Panic)
	Panic Panic
}

func (b *Block) catch() {
	if r := recover(); r != nil {
		b.Panic = r.(Panic)
		if b.Catch != nil {
			b.Catch(b.Panic)
		}
	}
}

// Do execute block.
func (b *Block) Do() {
	defer b.catch()
	b.Try()
}
