package mocks

//go:generate mockery -name DiffStore -dir ../../diff -output .
//go:generate mockery -name FailureStore -dir ../failurestore -output .
//go:generate mockery -name MetricsStore -dir ../metricsstore -output .
