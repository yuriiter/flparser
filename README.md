## Freelancer.com Project Scraper CLI (`flparser`)

`flparser` is a command-line tool written in Go that allows you to scrape public project listings from Freelancer.com based on various filtering criteria and export the results to **Markdown (.md)**, **CSV (.csv)**, or **JSON (.json)** formats.

### Installation

1.  **Prerequisites:** Ensure you have Go installed on your system.
2.  **Install via `go install`:**
    ```bash
    go install github.com/yuriiter/flparser@latest
    ```
    The executable will be placed in your `$GOPATH/bin` or `$GOBIN`.

---

### Usage & Parameters

The tool uses the following command-line flags to construct the search URL:

| Parameter | Flag | Default | Description |
| :--- | :--- | :--- | :--- |
| Project Types | `--types` | `hourly,fixed` | Filter by project type. Options: `hourly,fixed`, `hourly`, or `fixed`. |
| Client Countries | `--clientCountries` | `ca,au,no,de,se,ch,gb,us` | Comma-separated list of country codes (e.g., `us,uk,ca`). |
| Minimum Fixed Price | `--fixedMin` | `0` (Not set) | Minimum price for fixed-price projects. |
| Maximum Fixed Price | `--fixedMax` | `0` (Not set) | Maximum price for fixed-price projects. |
| Minimum Hourly Rate | `--hourlyMin` | `0` (Not set) | Minimum rate for hourly projects. |
| Maximum Hourly Rate | `--hourlyMax` | `0` (Not set) | Maximum rate for hourly projects. |
| Skills | `--skills` | `7,9,13,...` (Long list of programming languages) | Comma-separated list of skill IDs, or use `all` to remove the skill filter from the URL. |
| Sort Option | `--sort` | `latest` | How to sort the results. Options: `oldest`, `lowestPrice`, `highestPrice`, `fewestBids`, `mostBids`. |
| Search Query | `-q` | `""` (Not set) | A text term to search for (e.g., `golang parser`). |
| Page Number | `--page` | `1` (Not set) | The page number to scrape (each page has 20 projects). |
| Output File | `-O`, `--output` | `""` (Not set) | Specify a complete output filename (e.g., `results.json`). This overrides `-X`. |
| Output Extension | `-X`, `--extension` | `""` (Default to `md` and `csv`) | Specify the output format if `-O` is not used. Options: `md`, `csv`, `json`. |

### Default Output Behavior

If neither `-O` (output file) nor `-X` (output extension) is provided:
*   The tool defaults to generating **two** files: one `.md` and one `.csv`.
*   The filename format will be `freelancer.com_{HH-MM-SS_DD-MM-YYYY}.md` and `freelancer.com_{HH-MM-SS_DD-MM-YYYY}.csv`.

### Auto-Completion Setup

The `flparser` CLI supports shell auto-completion via `cobra`.

1.  **Generate the completion script for your shell:**
    *   **Bash (Linux/macOS):**
        ```bash
        flparser completion bash > flparser.bash
        sudo mv flparser.bash /etc/bash_completion.d/
        # You may need to restart your terminal or run 'source ~/.bashrc'
        ```
    *   **Zsh:**
        ```bash
        flparser completion zsh > _flparser
        # Move the script to one of the directories in your $fpath. E.g.:
        # mv _flparser ~/.zsh/completion/_flparser
        ```
    *   **PowerShell / Fish** are also supported (see `flparser completion --help`).

2.  **Usage:** After setup, type `flparser --` and press the `[Tab]` key to see available commands and flags.
