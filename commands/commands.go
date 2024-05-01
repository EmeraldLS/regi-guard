package commands

import (
	"fmt"
	"net/http"
)

// NewULID generates a new ULID
// Returns the ULID as a string
//
//	func NewULID() {
//		// Create a new ULID
//		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
//		ms := ulid.Timestamp(time.Now())
//		new_ulid, err := ulid.MustNew(ms, entropy)
//		if err != nil {
//			fmt.Println("Error while generating ULID: ", err)
//			return ""
//		}
//		return new_ulid.String()
//	}

type ErrResponse struct {
	STATUS_CODE     int    `json:"status_code,omitempty"`
	MESSAGE         string `json:"message,omitempty"`
	STATE_TO_ACCESS string `json:"state_to_access,omitempty"`
}

type SuccessResponse struct {
	STATUS_CODE int    `json:"status_code,omitempty"`
	MESSAGE     string `json:"message,omitempty"`
}

func NewSuccessResponse(msg string) *SuccessResponse {
	return &SuccessResponse{
		STATUS_CODE: http.StatusOK,
		MESSAGE:     msg,
	}
}

func NewErrResponse(state_to_access string, errMsg error) *ErrResponse {
	return &ErrResponse{
		STATUS_CODE:     http.StatusInternalServerError,
		MESSAGE:         errMsg.Error(),
		STATE_TO_ACCESS: state_to_access,
	}
}

type WSMessage struct {
	Type string      `json:"type,omitempty" validate:"required"`
	Data interface{} `json:"data,omitempty" validate:"required"`
}

type TinyTestInput struct {
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	Age       int      `json:"age,omitempty"`
	Address   Address  `json:"address"`
	Hobbies   []string `json:"hobbies"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode int    `json:"zip_code"`
}

type TinyTestOutput struct {
	FullName        string   `json:"full_name,omitempty"`
	Age             int      `json:"age,omitempty"`
	FullAddress     string   `json:"full_address,omitempty"`
	Hobbies         []string `json:"hobbies,omitempty"`
	NumberOfHobbies int      `json:"number_of_hobbies,omitempty"`
}

// start tiny test
func SerializeTinyTestInput(input *TinyTestInput) (*TinyTestOutput, error) {
	hobbyCount := tinyTestCountHobbies(input.Hobbies)
	full_name := tinyTestCreateFullName(*input)
	full_addr := tinyTestFlattenAddress(input.Address)

	output := &TinyTestOutput{
		FullName:        full_name,
		Age:             input.Age,
		FullAddress:     full_addr,
		Hobbies:         input.Hobbies,
		NumberOfHobbies: hobbyCount,
	}

	// 1.  create an instance of the data model for input and output
	// 2.  call commands on the instance of the data model to transform the data through the pipeline
	// create channel to receive the results

	// TinyTestCreateFullName(input, output) FullName {
	// 	// checks if output already has a value if it does no reason to re-compute
	// 	if output.Fullname != nil {
	// 		return nil, err
	// 	}
	// 	// checks if there was an error in tinyTestCreateFullName
	// 	if result != nil {
	// 		return nil, err
	// 	}
	// 	// Checks if Output.FullName is not equal to FullName . this allows us to avoid racing conditions because if its the same leave it
	// 	// if its not dont. but relative to actor id.
	// 	// error out if ulid is same
	// 	if output != FullName {
	// 		output.FullName = FullName
	// 	}
	// 	return nil, err
	// }

	// TinyTestFlattenAddress(input, output)
	// count number of hobies in the list
	return output, nil
}

func tinyTestCountHobbies(hobbies []string) int {
	return len(hobbies)
}

func tinyTestCreateFullName(input TinyTestInput) string {
	fullname := fmt.Sprintf("%s %s", input.FirstName, input.LastName)
	return fullname
}

func tinyTestFlattenAddress(addr Address) string {
	full_addr := fmt.Sprintf("%s %s, %s - %d", addr.Street, addr.City, addr.State, addr.ZipCode)
	return full_addr
}

// end tiny test

// actor manager can be thought of as a command manager keep track of number of running processes/routines workers etc and the state of the command

func SoftPullCommand() {
	panic("unimplemented")
}

// func SoftPullCommand(ULID *ulid.ULID, context *actor.Context) *actor.Actor {
// 	//1.  create an instance of the data model for input and output

// 	//2.  call commands on the instance of the data model to transform the data through the pipeline
// 	//3.  for each new actor or command we can run in parralel or in sequence but in this scenairo we want a channel
// 	//4.  to receive the results of the go routines that are processing the data as commands 'aka' actors
// 	//5.  finnally mark this command as complete and return the results to the caller aka the instance of the response model

// 	// upgrades after the first version to limit the number of go routines that can be run at once by using the actor manager aka the 'channel'
// 	// create a new actor
// 	// create a channel to receive the results
// 	// spin off several go routines to process the data in parallel aka actors
// 	// wait for all the go routines to finish
// 	// combine the results

// 	return nil
// }

func CalculateFicoScore() {
	panic("unimplemented")
}
