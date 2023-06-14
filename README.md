# secret-edit

Uma ferramenta de linha de comando para editar segredos armazenados no Google Cloud Secret Manager.

## Requisitos:

- Go 1.19 ou superior
- GCP SDK

## Como funciona ?

A ferramenta irá se autenticar através de uma service account ou das credenciais padrões da gcp criada através do comando: `gcloud auth application-default login`

Depois da autenticação ela irá procurar pela secret na qual você informou presente no projeto X, caso ela exista será aberto o editor de texto padrão do seu sistema operacional com o conteúdo da secret em formato YAML, nesse processo é criado um arquivo temporário em `/tmp` seguindo o padrão do nome do arquivo: scts-<hash>.yaml, após a edição e o salvamento do arquivo a ferramenta irá criar uma nova versão da secret com o conteúdo editado em formato JSON, o arquivo temporário logo é apagado automaticamente.

## Instalação

Esta ferramenta requer o Go instalado em sua máquina. Você pode instalar o Go seguindo as instruções em golang.org/doc/install.

Para instalar o secret-edit, execute o seguinte comando:

Execute o comando abaixo para compilar o binário:

```bash
make && make install
```

### Uso

Para editar um segredo, execute o seguinte comando:

```bash
secret-edit <secret-name> --project <project-id>
```

Substitua <nome-do-segredo> pelo ID do segredo que deseja editar.

O comando edit abrirá seu editor padrão de acordo com a variável de ambiente EDITOR e permitirá que você edite o conteúdo do segredo no formato YAML. Quando você salvar e sair do editor, a ferramenta criará uma nova versão do segredo com o conteúdo editado.

Para listar todos os segredos em um projeto, execute o seguinte comando:

```bash
secret-edit --list --project <project-id>
```

Substitua <id-do-projeto> pelo ID do projeto do GCP que contém os segredos.
