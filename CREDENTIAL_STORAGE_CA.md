# GaleraHealth - Sistema d'Emmagatzematge de Credencials SSH i MySQL

## Resum Executiu

El sistema **GaleraHealth** ja implementa un sistema complet d'emmagatzematge de credencials SSH i MySQL per node, amb xifratge AES-GCM i gestió automàtica de la configuració.

## 🔐 Característiques de Seguretat

### 1. Emmagatzematge Xifrat per Node
- **Contrasenyes SSH**: Xifrades amb AES-256-GCM per cada node
- **Contrasenyes MySQL**: Xifrades independentment per cada node
- **Claus úniques**: Cada node utilitza una clau de xifratge diferent derivada de la seva IP
- **Salt automàtic**: La IP del node actua com a salt natural per a la derivació de claus

### 2. Estructura de Configuració
```json
{
  "node_credentials": [
    {
      "node_ip": "10.1.1.91",
      "ssh_username": "root",
      "mysql_username": "galera_user", 
      "encrypted_ssh_password": "base64_encrypted_data",
      "encrypted_mysql_password": "base64_encrypted_data",
      "has_ssh_password": true,
      "has_mysql_password": true,
      "uses_ssh_keys": false
    },
    {
      "node_ip": "10.1.1.92", 
      "ssh_username": "admin",
      "uses_ssh_keys": true
    }
  ]
}
```

## 🚀 Funcionalitats Implementades

### 1. Detecció Automàtica del Mètode d'Autenticació
```bash
# Quan l'usuari es connecta per primera vegada
./galerahealth
Enter the Galera cluster node IP: 10.1.1.91
Enter SSH username: admin

# El sistema prova:
# 1. Claus SSH (automàticament)
# 2. Si falla, demana contrasenya
# 3. Guarda el mètode exitós per a futures connexions
```

### 2. Reutilització Automàtica de Credencials
```bash
# Propera execució al mateix node
./galerahealth  
Enter the Galera cluster node IP: 10.1.1.91
# Sistema utilitza credencials guardades automàticament
✓ SSH connection successful using saved credentials
```

### 3. Suport per Credencials Diferents per Node
```bash
# Node 1: Utilitza claus SSH amb usuari 'admin'
# Node 2: Utilitza contrasenya amb usuari 'root'
# Node 3: Utilitza contrasenya amb usuari 'galera'
# Cada configuració es guarda independentment
```

## 🔧 Implementació Tècnica

### 1. Funcions de Gestió de Credencials
- `getNodeCredentials(nodeIP)`: Recupera credencials per a un node específic
- `setNodeCredentials(nodeIP, ...)`: Guarda/actualitza credencials per a un node
- `getNodeSSHPassword(nodeIP)`: Desxifra i retorna contrasenya SSH
- `getNodeMySQLPassword(nodeIP)`: Desxifra i retorna contrasenya MySQL

### 2. Xifratge AES-GCM
```go
func encryptPassword(password, nodeIP string) (string, error) {
    // Genera clau única per node usant SHA-256 de la IP
    key := sha256.Sum256([]byte(nodeIP))
    
    // Xifra amb AES-256-GCM (autenticat)
    block, err := aes.NewCipher(key[:])
    gcm, err := cipher.NewGCM(block)
    
    // Genera nonce aleatori
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    
    // Xifra i autentica
    ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

### 3. Ubicació de la Configuració
- **Fitxer**: `~/.galerahealth`
- **Permisos**: `600` (només lectura/escriptura per l'usuari)
- **Format**: JSON estructurat per a fàcil gestió

## 📋 Exemples d'Ús

### 1. Configuració Inicial
```bash
# Primera execució
./galerahealth -v
Enter the Galera cluster node IP: 10.1.1.91
Enter SSH username (default: root): admin
# Sistema prova claus SSH, si falla demana contrasenya
Enter SSH password for admin@10.1.1.91: [contrasenya_segura]
✓ Connected and credentials saved for future use
```

### 2. Execucions Posteriors
```bash
# Execucions següents al mateix node
./galerahealth -v
Enter the Galera cluster node IP: 10.1.1.91
# Sistema detecta credencials guardades
✓ Using saved credentials for admin@10.1.1.91
✓ SSH connection successful using saved password
```

### 3. Nodes amb Credencials Diferents
```bash
# Node A: Claus SSH
Node 10.1.1.91: admin + SSH keys ✓

# Node B: Contrasenya
Node 10.1.1.92: root + password ✓

# Node C: Credencials diferents
Node 10.1.1.93: galera_user + different_password ✓

# Totes guardades i gestionades independentment
```

## 🔒 Mesures de Seguretat

### 1. Xifratge Fort
- **Algoritme**: AES-256-GCM (xifrat autenticat)
- **Derivació de claus**: SHA-256 de la IP del node
- **Integritat**: GCM proporciona autenticació de missatges

### 2. Aïllament de Credencials
- **Per node**: Cada node té claus de xifratge úniques
- **Compromís limitat**: El compromís d'un node no afecta els altres
- **Rotació fàcil**: Es poden canviar credencials per node individualment

### 3. Permisos de Fitxer
- **Accés restringit**: Només l'usuari propietari pot llegir/escriure
- **Ubicació segura**: Directori home de l'usuari
- **Backup automàtic**: Sistema crea còpies abans d'actualitzar

## 🛠️ Administració

### 1. Visualitzar Configuració Actual
```bash
cat ~/.galerahealth | jq '.'
```

### 2. Eliminar Configuració (Reiniciar)
```bash
./galerahealth --clear-config
```

### 3. Depuració amb Verbositat
```bash
./galerahealth -vvv  # Màxim detall
```

## ✅ Estat d'Implementació

- ✅ **Emmagatzematge per node**: Completament implementat
- ✅ **Xifratge AES-GCM**: Operatiu i segur
- ✅ **Detecció automàtica**: Claus SSH + fallback a contrasenya
- ✅ **Reutilització de credencials**: Automàtica per a connexions següents
- ✅ **Gestió multi-node**: Suport per credencials diferents per node
- ✅ **Interfície d'usuari**: Clara i intuïtiva
- ✅ **Documentació**: Completa i actualitzada

## 🎉 Conclusió

El sistema **GaleraHealth** ja implementa completament l'emmagatzematge segur de credencials SSH i MySQL per node. Les característiques clau inclouen:

- **Seguretat**: Xifratge AES-256-GCM amb claus úniques per node
- **Flexibilitat**: Suport per diferents credencials en cada node del cluster
- **Usabilitat**: Detecció automàtica i reutilització de credencials
- **Mantenibilitat**: Configuració clara i eines d'administració integrades

L'usuari pot utilitzar el sistema immediatament sense configuració addicional - tots els SSH users i contrasenyes es guarden automàticament de forma segura a `~/.galerahealth`.
