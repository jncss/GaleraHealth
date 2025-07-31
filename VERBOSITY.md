# GaleraHealth - Sistema de Verbositat

## Descripció General

GaleraHealth ara inclou un sistema de verbositat amb tres nivells diferents que permet als usuaris controlar la quantitat d'informació mostrada durant l'execució.

## Nivells de Verbositat

### Nivell 0 - Mínim (Per defecte)
**Ús:** `./galerahealth` (sense flags)

**Què mostra:**
- Només missatges essencials i resultats finals
- Conexions exitoses
- Títols de les seccions principals
- Errors crítics

**Exemple de sortida:**
```
=== GaleraHealth - Galera Cluster Monitor ===
✓ Successfully connected to node 10.1.1.91
=== GALERA CLUSTER INFORMATION ===
🔍 Performing cluster coherence analysis...
```

### Nivell 1 - Normal (-v)
**Ús:** `./galerahealth -v`

**Què mostra:**
- Tot del nivell mínim +
- Carrega de configuració guardada
- Advertències i avisos
- Confirmacions d'operacions

**Exemple de sortida:**
```
=== GaleraHealth - Galera Cluster Monitor ===
📋 💾 Loaded saved configuration from /home/user/.galerahealth
📋 ⚠️  Connection with keys failed: permission denied
📋 🔐 Attempting connection with password...
✓ Successfully connected to node 10.1.1.91
```

### Nivell 2 - Verbose (-vv)
**Ús:** `./galerahealth -vv`

**Què mostra:**
- Tot dels nivells anteriors +
- Detalls de les operacions internes
- Informació de cerca de fitxers
- Processos de connexió SSH detallats
- Anàlisi de configuració paso a paso

**Exemple de sortida:**
```
🔍 Verbosity level set to: 2
=== GaleraHealth - Galera Cluster Monitor ===
🔍 Application started with verbosity level 2
📋 💾 Loaded saved configuration from /home/user/.galerahealth
🔍    Last used: Node IP: 10.1.1.91, SSH User: root, MySQL User: root
🔍 Attempting SSH connection to root@10.1.1.91
🔍 Gathering cluster information from initial node
📋 🔍 Searching for cluster information...
🔍 📁 Searching for configuration files...
🔍 📁 Configuration files found: 3 files
```

### Nivell 3 - Debug (-vvv)
**Ús:** `./galerahealth -vvv`

**Què mostra:**
- Tot dels nivells anteriors +
- Informació completa de depuració
- Detalls de xifrat/desxifrat de contrasenyes
- Llistes completes de fitxers trobats
- Configuracions internes
- Dades en brut

**Exemple de sortida:**
```
🐛 Verbosity level set to: 3
=== GaleraHealth - Galera Cluster Monitor ===
🐛 Application started with verbosity level 3
📋 💾 Loaded saved configuration from /home/user/.galerahealth
🔍    Last used: Node IP: 10.1.1.91, SSH User: root, MySQL User: root
🐛 Updated configuration: NodeIP=10.1.1.91, Username=root, CheckCoherence=true
🐛 CheckMySQL set to: true
🔍 Found saved encrypted password
🐛 Attempting to decrypt stored password
🐛 Password successfully decrypted
🐛   - /etc/mysql/conf.d/galera.cnf
🐛   - /etc/mysql/mysql.conf.d/mysqld.cnf
```

## Formats de Flags Suportats

El sistema suporta diferents formats per especificar el nivell de verbositat:

```bash
./galerahealth -v      # Nivell 1 (normal)
./galerahealth -vv     # Nivell 2 (verbose)
./galerahealth -vvv    # Nivell 3 (debug)
```

## Icones Utilitzades per Nivell

- **📋** - Missatges normals (-v i superior)
- **🔍** - Informació detallada (-vv i superior)  
- **🐛** - Informació de depuració (-vvv només)

## Compatibilitat amb Altres Flags

Els flags de verbositat es poden combinar amb altres opcions:

```bash
./galerahealth -vv --clear-config    # Neteja amb verbositat detallada
./galerahealth -v --help             # Ajuda (la verbositat no afecta l'ajuda)
```

## Quan Utilitzar Cada Nivell

- **Mínim (per defecte):** Ús normal diari, només vols veure els resultats
- **Normal (-v):** Quan vols més informació sobre què està passant
- **Verbose (-vv):** Per resolucionar problemes o entendre els processos interns
- **Debug (-vvv):** Per desenvolupament o diagnòstic profund de problemes

## Implementació Tècnica

El sistema utilitza funcions de logging centralitzades:
- `logMinimal()` - Sempre es mostra
- `logNormal()` - Només amb -v i superior
- `logVerbose()` - Només amb -vv i superior
- `logDebug()` - Només amb -vvv

Això permet un control granular de la sortida sense afectar el rendiment quan no es necessita informació detallada.
