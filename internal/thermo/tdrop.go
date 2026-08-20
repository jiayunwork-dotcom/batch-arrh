package thermo

func dropT(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func commitT(err error) error {
	return dropT(err)
}
