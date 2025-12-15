# Sun8x8 – Generator czcionek 8x8 dla matryc LED

**Autor:** SunRiver / Lothar TeaM  
**Forum:** [https://forum.lothar-team.pl/](https://forum.lothar-team.pl/)  
**Wersja:** 0.0.3

<img width="554" height="1066" alt="Zrzut ekranu 2025-12-15 174100" src="https://github.com/user-attachments/assets/758a5525-8eeb-4f19-af16-933b571560f1" />

<img width="732" height="664" alt="Zrzut ekranu 2025-12-12 220012" src="https://github.com/user-attachments/assets/4d539142-9518-4bc4-839c-5c4b6017d37b" />


---

## Opis

Sun8x8 to edytor i generator czcionek 8x8 przeznaczony do matryc LED, napisany w Go z użyciem biblioteki [Ebiten](https://ebiten.org/).  
Program umożliwia:

- Edycję pikseli w trybach:
  - 1-bit (MONO)
  - 2-bit (2 kolory na piksel)
  - RGB565
- Eksport do:
  - C (czysty)
  - PROGMEM (Arduino / AVR)
  - PNG
  - JSON (kolory w formacie HEX)
- Podgląd w formacie HEX i BIN
- Animowany podgląd przewijający 8x8 glyphy w symulowanej szerokiej matrycy
- Suwak regulujący prędkość animacji

---

## Wymagania

- Go 1.24+  
- Biblioteka [Ebiten v2](https://github.com/hajimehoshi/ebiten)  
- Czcionka TrueType (`resource/Itim-Regular.ttf`)

---

## Instrukcja użytkowania

Kliknięcie komórki – zmienia stan piksela.<br>
M – zmiana trybu edycji (MONO / 2-KOLORY / RGB)<br>
C – wyczyść matrycę<br>
<bvr>
Przyciski po prawej – eksport do C, PROGMEM, PNG, podgląd HEX/BIN, start/stop animacji.<br>
Suwak pod panelem – kontrola prędkości animacji.<br>
<br>
<br>

## Instalacja

1. Sklonuj repozytorium:

```bash
git clone https://github.com/SunDUINO/Sun8x8_Font_generator.git
cd Sun8x8_Font_generator

```

2. Pobierz zależności:
```bash
go mod tidy
```

3. Uruchom program:
```bash
go run .
```
