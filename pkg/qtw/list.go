package qtw

import (
	qt "github.com/mappu/miqt/qt6"
)

type ListWrapper struct {
	lw         *qt.QListWidget
	labelIndex map[string]int
	indexLabel []string
	curIdx     int
}

func List() *ListWrapper {
	return &ListWrapper{lw: qt.NewQListWidget2(), labelIndex: make(map[string]int)}
}

func (l *ListWrapper) Item(label string) *ListWrapper {
	l.lw.AddItem(label)
	l.labelIndex[label] = l.curIdx
	l.indexLabel = append(l.indexLabel, label)
	l.curIdx++
	return l
}

func (l *ListWrapper) Frame(f qt.QFrame__Shape) *ListWrapper {
	l.lw.SetFrameShape(f)
	return l
}

func (l *ListWrapper) Widget() *qt.QListWidget {
	return l.lw
}

func (l *ListWrapper) Set(label string) {
	if i, ok := l.labelIndex[label]; ok {
		l.lw.SetCurrentRow(i)
	}
}

func (l *ListWrapper) OnClick(f func(label string)) *ListWrapper {
	l.lw.OnClicked(func(index *qt.QModelIndex) {
		f(l.indexLabel[index.Row()])
	})
	return l
}
