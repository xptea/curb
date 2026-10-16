"""Repository paths and tool versions shared by the bundler."""
import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[2]
TOOLS = ROOT / "tools"
CACHE = ROOT / "build" / "downloads"
VERSION = json.loads((ROOT / "mygo.json").read_text())["version"]
TOOL_VERSIONS = json.loads((ROOT / "packaging" / "tools.json").read_text())
MYGO_VERSION = re.search(r"github.com/egoist/mygo\s+(v\S+)", (ROOT / "go.mod").read_text())[1]
HEADERS = {"User-Agent": f"Curb-build/{VERSION}", "Accept": "application/vnd.github+json"}
