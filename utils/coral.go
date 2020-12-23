package utils

func (coral *Coral) valCommands() bool{
	if coral.Commands.Prefix == ""{ return false }
	if coral.Commands.Command == ""{ return false }
	if coral.Commands.Message == ""{ return false }
	return true
}

func (coral *Coral) valSchedule() bool{
	if coral.Schedule.Rule == ""{ return false }
	if coral.Schedule.Sender == ""{ return false }
	if coral.Schedule.Message == ""{ return false }
	return true
}

func (coral *Coral) valGreeting() bool{
	if coral.DefaultGreeting.Message == ""{ return false }
	return true
}

func (coral *Coral) valAuthor() bool{
	if coral.Author.Name == "" { return false }
	if coral.Author.Phone == "" { return false }
	if coral.Author.Email == "" { return false }
	return true
}

func (coral *Coral) CommandExist() bool{
	if coral.Commands.Prefix == "" || coral.Commands.Command == "" || coral.Commands.Message == ""{ 
		return false 
	}
	return true
}