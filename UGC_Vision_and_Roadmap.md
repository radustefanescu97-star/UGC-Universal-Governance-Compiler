# UGC (Universal Governance Compiler) - Vision & Roadmap

## Viziunea de Ansamblu
Proiectul UGC nu este doar un instrument tehnic (un compilator), ci un **vehicul pentru standardizarea modului în care agenții AI interacționează cu codul**. Acesta rezolvă două probleme majore:
1. **Problema tehnică (Mecanismul):** Fragmentarea masivă a configurațiilor agenților AI (Cursor, Antigravity, Copilot, Windsurf).
2. **Problema metodologică (Filosofia):** Lipsa de disciplină, siguranță și structură în lucrul cu agenții AI, care duce la cod haotic și riscuri de execuție.

Pentru a avea succes, UGC va fi livrat cu o filosofie implicită de bune practici, inspirată puternic din guvernanța matură **PromSpace**.

## Cele Două Pilastru Ale Proiectului

### 1. Standardul Universal (Filosofia / Metodologia)
UGC definește cum ar trebui să arate o structură de guvernanță agnostică, de tip *Single Source of Truth*. 
La inițializarea proiectului (`ugc init`), developerul primește nu doar un folder gol, ci un set de template-uri bazate pe disciplina PromSpace:
- **Mandatory Bootstrap:** Regula prin care agentul trebuie să citească un context specific (SOP-uri) înainte de a face un plan material sau de a acționa.
- **Worklog Discipline:** Obligativitatea documentării sesiunilor, pașilor și acțiunilor luate de AI.
- **Approval Gates:** Puncte de oprire clare, în care acțiunile distructive, de cloud sau care generează costuri au nevoie de acordul explicit al unui om.
- **Target & Protected Surfaces:** Delimitarea explicită a zonelor de cod pe care agentul are voie să le modifice, protejând restul sistemului.
- **Sistem de Actualizare:** Standardul nu este static. UGC va include un mecanism de update (ex. prin comanda `ugc update`) care permite preluarea celor mai noi versiuni de bune practici și SOP-uri din sursa centrală publică, fără a suprascrie personalizările specifice făcute de echipa care deține proiectul.

### 2. Compilatorul (Mecanismul Software)
Software-ul în sine este un utilitar CLI lightweight și rapid (ideal scris în **Go** pentru distribuție zero-dependency pe orice sistem de operare).
Rolul lui este să preia "Standardul Universal" și să-l traducă în dialectele stricte specifice fiecărui IDE / Agent.

#### Arhitectura de Bază:
- **CLI (`cli/`):** Gestionează interfața cu utilizatorul:
  - `ugc init [--analyze]`: Creează structura de bază și opțional scanează repository-ul (`go.mod`, `package.json`) pentru a infera constrângerile tehnice (Auto-Discovery).
  - `ugc build`: Compilează regulile din `.universal-governance/`.
  - `ugc audit`: Verifică dacă fișierele emise (ex. `.cursorrules`) au fost alterate manual, detectând astfel devierile de la guvernanță (Drift Detection).
- **Core Engine (`engine/`):** Parsează regulile agnostice din folderul `.universal-governance/` și le transpune într-un model intern unificat în memorie.
- **Emitters (`emitters/`):** Modulul de "Traducători". Pachete independente care preiau modelul intern și generează output-ul final:
  - `emitter-cursor` -> concatenează și generează `.cursorrules`
  - `emitter-antigravity` -> generează structura de foldere `.agents/AGENTS.md` și injectează skill-uri (ex: pentru worklog).
  - `emitter-copilot` -> generează `.github/copilot-instructions.md`

### 3. Platforme AI Suportate (Listă Oficială și Închisă)
Pentru a oferi transparență și claritate, dar mai ales pentru a **garanta** că regulile critice de guvernanță nu sunt "pierdute în traducere" de un ecosistem haotic de plugin-uri, proiectul menține o listă strictă. Pentru faza inițială, vom suporta exclusiv:
1. **OpenAI Codex**
2. **Google Antigravity 2.0**
3. **Claude Code**
4. **Cursor (IDE)**

*(Notă despre **Cursor**: Spre deosebire de agenții CLI, Cursor este un IDE complet cu AI integrat, care funcționează ghidându-se după un fișier local `.cursorrules` generat de UGC).*

## Traseul Utilizatorului (The Developer Journey)

Acesta este fluxul complet prin care un dezvoltator de pe propriul laptop va adopta și utiliza UGC:

1. **Instalarea și Inițializarea**
   - Developerul instalează utilitarul UGC la nivel de sistem (via release-uri de GitHub, Homebrew etc.).
   - Rulează comanda `ugc init` în rădăcina proiectului său de cod.
   - UGC creează folderul `.universal-governance/` și îl populează cu template-urile de bază (metodologia PromSpace generalizată).

2. **Personalizarea Guvernanței**
   - Dezvoltatorul și echipa sa ajustează fișierele Markdown pentru a reflecta realitatea proiectului lor (ex: ce framework-uri folosesc, cine dă aprobările, care este parola de aprobare, ce foldere sunt esențiale).

3. **Compilarea / Build-ul**
   - Se rulează comanda `ugc build` (poate fi rulată manual sau automat printr-un git hook / script CI/CD).
   - Core Engine-ul procesează toate regulile.
   - Emitters generează/suprascriu fișierele vendor-specific în repository (`.cursorrules`, etc.).

4. **Execuția cu Agenți AI (Fără Frecare)**
   - Developerul deschide Cursor. Cursor citește `.cursorrules` generat și aplică regulile echipei.
   - Același developer, sau un alt coleg din echipă, pornește un agent Google Antigravity. Antigravity citește `.agents/AGENTS.md` generat de UGC și **respectă absolut aceleași reguli de guvernanță**.
   - Nu există discrepanțe sau scăpări de complianță între agenți.

## Concluzie
UGC este mai mult decât un simplu parser sau convertor de Markdown. 
Din perspectiva business/comunității, este un **"Trojan Horse" educațional**. Sub pretextul rezolvării fragmentării tehnice, UGC injectează în comunitatea open-source o modalitate matură, sigură și extrem de disciplinată de a colabora cu inteligența artificială.
