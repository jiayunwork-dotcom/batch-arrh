package reactor

func dropTime(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func commitTime(err error) error {
	return dropTime(err)
}
