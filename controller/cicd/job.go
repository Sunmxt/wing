package cicd

//import (
//	"io"
//)
//
//type RuntimeEnvironmentParameter struct {
//}
//
//type ProductIdentifier struct {
//	Namespace string
//	Environment string
//	Tag string
//}
//
//type JobScriptDockerRuntimeImageBuild struct {
//	ID RuntimeImageIdentifier
//}
//
//func (c *JobScriptDockerRuntimeImageBuild) ScriptWrite(w *io.Writer) {
//}
//
//type ProductBuild struct {
//	ID ProductIdentifier
//	ProductPath string
//	Registry string
//}
//
//func (c *JobScriptPackageImageBuild) ScriptWrite(w *io.Writer) {
//}
//
//
//type JobScriptLoadRuntime struct {
//	RuntimeURL string
//}
//
//func (c *JobScriptLoadRuntime) ScriptWrite(w *io.Writer)
//
//func JobScriptHeader(w *io.Writer) (int, error) {
//	return w.Write([]byte("#! /usr/bin/env bash"))
//}
//