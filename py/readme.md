# Requirements

```text
# empty
```

```powershell
# reinstall
python -m venv venv\
venv\Scripts\activate
python -m pip install --upgrade pip setuptools wheel
# python -m pip install -r requirements.txt
python --version >requirements-installed-venv.txt
python -m pip list >>requirements-installed-venv.txt

# activate
venv\Scripts\activate
# work ... then
deactivate
```

```bash
# with pyenv
deactivate
pyenv local 3.14.3 # creates a .python-version file

# recreate
rm -rf .venv && python -m venv .venv/
.venv/bin/python --version
.venv/bin/pip install --upgrade pip setuptools wheel
#.venv/bin/pip install -r requirements.txt
.venv/bin/python --version >requirements-installed-venv.txt
.venv/bin/pip list >>requirements-installed-venv.txt
cat requirements-installed-venv.txt

# activate
source .venv/bin/activate
# work ... then
deactivate

# requirements-installed-venv.txt
Python 3.14.3
Package    Version
---------- -------
packaging  26.0
pip        26.0.1
setuptools 82.0.1
wheel      0.46.3
```
