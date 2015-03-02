/* walter: a deployment pipeline template
 * Copyright (C) 2014 Recruit Technologies Co., Ltd. and contributors
 * (see CONTRIBUTORS.md)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package stages

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/recruit-tech/walter/log"
)

type CommandStage struct {
	BaseStage
	Command   string `config:"command"`
	Directory string `config:"directory"`
	OnlyIf    string `config:"only_if"`
	OutResult string
	ErrResult string
}

func (self *CommandStage) GetStdoutResult() string {
	return self.OutResult
}

func (self *CommandStage) Run() bool {
	// Check OnlyIf
	if self.runOnlyIf() == false {
		log.Warnf("[command] exec: skipped stage \"%s\", since only_if condition failed", self.BaseStage.StageName)
		return false
	}

	// Run command
	return self.runCommand()
}

func (self *CommandStage) runOnlyIf() bool {
	if self.OnlyIf == "" {
		log.Infof("[command] only_if: %s stage does not have \"only_if\" attribute", self.BaseStage.StageName)
		return true
	}

	cmd := exec.Command("sh", "-c", self.OnlyIf)
	log.Infof("[command] only_if: %s", self.BaseStage.StageName)
	log.Debugf("[command] only_if literal: %s", self.OnlyIf)
	cmd.Dir = self.Directory
	out, err := cmd.StdoutPipe()
	outE, errE := cmd.StderrPipe()

	if err != nil {
		log.Errorf("[command] only_if err: %s", out)
		log.Errorf("[command] only_if err: %s", err)
		return false
	}

	if errE != nil {
		log.Errorf("[command] only_if err: %s", outE)
		log.Errorf("[command] only_if err: %s", errE)
		return false
	}

	err = cmd.Start()
	if err != nil {
		log.Errorf("[command] only_if err: %s", err)
		return false
	}
	self.OutResult = copyStream(out, "only_if")
	self.ErrResult = copyStream(outE, "only_if")

	err = cmd.Wait()
	if err != nil {
		log.Errorf("[command] only_if err: %s", err)
		return false
	}
	return true
}

func (self *CommandStage) runCommand() bool {
	cmd := exec.Command("sh", "-c", self.Command)
	log.Infof("[command] exec: %s", self.BaseStage.StageName)
	log.Debugf("[command] exec command literal: %s", self.Command)
	cmd.Dir = self.Directory
	out, err := cmd.StdoutPipe()
	outE, errE := cmd.StderrPipe()

	if err != nil {
		log.Errorf("[command] exec err: %s", out)
		log.Errorf("[command] exec err: %s", err)
		return false
	}

	if errE != nil {
		log.Errorf("[command] exec err: %s", outE)
		log.Errorf("[command] exec err: %s", errE)
		return false
	}

	err = cmd.Start()
	if err != nil {
		log.Errorf("[command] exec err: %s", err)
		return false
	}
	self.OutResult = copyStream(out, "exec")
	self.ErrResult = copyStream(outE, "exec")

	err = cmd.Wait()
	if err != nil {
		log.Errorf("[command] exec err: %s", err)
		return false
	}
	return true
}

func copyStream(reader io.Reader, prefix string) string {
	var err error
	var n int
	var buffer bytes.Buffer
	tmpBuf := make([]byte, 1024)
	for {
		if n, err = reader.Read(tmpBuf); err != nil {
			break
		}
		buffer.Write(tmpBuf[0:n])
		log.Infof("[command] %s output: %s", prefix, tmpBuf[0:n])
	}
	if err == io.EOF {
		err = nil
	} else {
		log.Error("ERROR: " + err.Error())
	}
	return buffer.String()
}

func (self *CommandStage) AddCommand(command string) {
	self.Command = command
	self.BaseStage.Runner = self
}

func (self *CommandStage) SetDirectory(directory string) {
	self.Directory = directory
}

func NewCommandStage() *CommandStage {
	stage := CommandStage{Directory: "."}
	return &stage
}
