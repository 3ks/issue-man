package tools

import "issue-man/global"

func (g getFunctions) String(source string) *string {
	return &source
}

func (g getFunctions) Strings(source []string) *[]string {
	newSlice := make([]string, len(source))
	copy(source, newSlice)
	return &newSlice
}

func (g getFunctions) Int(source int) *int {
	if source == 0 {
		return nil
	}
	return &source
}

func (g getFunctions) Float(source float64) *float64 {
	if source == 0 {
		return nil
	}
	return &source
}

// 只是一个简写
func (g getFunctions) SourceOwnerAndRepository() string {
	return global.Conf.Repository.Spec.Source.Owner
}

// 只是一个简写
func (g getFunctions) SourceRepository() string {
	return global.Conf.Repository.Spec.Source.Repository
}

// 只是一个简写
func (g getFunctions) WorkspaceOwner() string {
	return global.Conf.Repository.Spec.Workspace.Owner
}

// 只是一个简写
func (g getFunctions) WorkspaceRepository() string {
	return global.Conf.Repository.Spec.Workspace.Repository
}
