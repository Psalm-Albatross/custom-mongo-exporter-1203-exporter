# 🧑🏿‍💻 Install and setup chocolatey in windows 11.

In this section we will learn how to install chocolatey package manager
for windows.

Before installing it let's try to know what is **Chocolatey?**

**Chocolatey** is a software management solution for Windows OS.

## Requirements

Let's check the requirement first to install **Chocolatey**

1. Windows 7+/Windows Server 2003+.
2. Powershell V2+.
3. .NET Framework 4+ (If you don't have *.NET Framework* installed it will try to install 
it to met the requirement).

📒: All below steps must be performed on powershell

## Steps to check windows os version

- Press 🪟 + <kbd>R</kbd> and then type <b>winver</b>
## Check powershell version

- Open command prompt CMD and type <b>powershell.exe</b> and then run <b>$PSVersionTable.PSVersion</b>

## Before installing choco check execution policy in windows

👉 Fore more information related to windows execution policy see official doc here: [about_Execution_Policies](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies?view=powershell-7.3)

- ```Get-ExecutionPolicy``` (If Restricted) then follow next step.
- ```Set-ExecutionPolicy AllSigned``` (Will set to AllSigned), We can Bypass Execution Policy but AllSigned is more secure.
- For Bypass run this ```Set-ExecutionPolicy Bypass -Scope Process```

# Install choco 
If you are done with setting up execution policy then procced with **chocolatey** installation.

```
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

If there is no error during installation then you have successfully installed choco

Next check if choco installed or not you do this by using below command(Run in CMD)

```
where choco
```
Above command will display the binary or .exe file of choco

Next try to run choco if above command will return the location of .exe file in windows.

```
choco
```

Above command will return installed version of chocolatey

For help page run below command

```
choco -?
```
👉 Follow official documentation for more info: [Install Choco on windows](https://chocolatey.org/install)

























