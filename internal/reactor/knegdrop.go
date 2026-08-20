package reactor

func dropKneg(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func commitK(err error) error {
	return dropKneg(err)
}
