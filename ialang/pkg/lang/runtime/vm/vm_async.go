package vm

func (v *VM) execAwait() error {
	val, err := v.pop()
	if err != nil {
		return err
	}
	resolved, awaitErr := v.asyncRuntime.AwaitValue(val)
	if awaitErr != nil {
		return awaitErr
	}
	v.push(resolved)
	return nil
}

