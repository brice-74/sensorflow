package clickhouse

func HandleError(err error, markDown func()) error {
	if err == nil {
		return nil
	}

	return err
}
