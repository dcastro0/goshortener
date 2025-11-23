package config

import (
	"log"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Aviso: Não foi possível encontrar o arquivo .env, usando variáveis de ambiente padrão. Erro: %v", err)
	}
}
