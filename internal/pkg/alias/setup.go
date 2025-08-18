package alias

func Setup() error {
	if err := CreateAliasFolder(); err != nil {
		return err
	}
	if err := CreateAliasLoaderScript(); err != nil {
		return err
	}
	if err := AddAliasLoaderScriptToZshrc(); err != nil {
		return err
	}
	return nil
}
