package pvz

import "github.com/prometheus/client_golang/prometheus"

func init() {
	prometheus.MustRegister(pvzCounter)
	prometheus.MustRegister(receptionsCounter)
	prometheus.MustRegister(productsCounter)
}

var pvzCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "pvzs_created_counter",
	Help: "Quantity of created pvzs by city",
}, []string{"city"})

func observeCreatedPVZ(city string) {
	pvzCounter.WithLabelValues(city).Inc()
}

var receptionsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "receptions_created_counter",
	Help: "Quantity of created receptions by pvzId",
}, []string{"pvzId"})

func observeCreatedReceptions(pvzId string) {
	receptionsCounter.With(prometheus.Labels{"pvzId": pvzId}).Inc()
}

var productsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "products_created_counter",
	Help: "Quantity of added products by product type",
}, []string{"productType"})

func observeCreatedProducts(productType string) {
	productsCounter.With(prometheus.Labels{"productType": productType}).Inc()
}
