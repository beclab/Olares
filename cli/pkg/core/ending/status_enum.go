/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package ending

type ResultStatus int

const (
	NULL    ResultStatus = -99
	SKIPPED ResultStatus = iota - 2
	SUCCESS
	FAILED
)

var EnumList = []ResultStatus{
	NULL,
	SKIPPED,
	SUCCESS,
	FAILED,
}

func (r ResultStatus) String() string {
	switch r {
	case SUCCESS:
		return "success"
	case FAILED:
		return "failed"
	case SKIPPED:
		return "skipped"
	case NULL:
		return "null"
	default:
		return "invalid option"
	}
}

func GetByCode(code int) ResultStatus {
	switch code {
	case -99:
		return NULL
	case -1:
		return SKIPPED
	case 0:
		return SUCCESS
	default:
		return FAILED
	}
}
