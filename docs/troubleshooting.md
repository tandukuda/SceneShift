# Troubleshooting

## "Access Denied" Errors
SceneShift interacts with system processes, which requires high-level permissions.
* **Solution:** Ensure you grant the **Administrator** permission request when launching the application. If you clicked "No", restart the app.

## Config Files Not Appearing
If `config.yaml` or `theme.yaml` are not generated:
* **Solution:** Check if you are running the `.exe` from a protected folder (like `Program Files`). Move the executable to a user folder like `Documents/SceneShift` or your Desktop.

## App Not Restarting (Restore Mode)
If pressing `R` does not reopen an app:
* **Solution:** The app might be missing its `exec_path`. Use the **Edit (`e`)** function in the UI to manually add the file path to the executable (e.g., `C:\Program Files\Adobe\Photoshop.exe`).

## App Not Closing (Quit Mode)
If pressing `Q` does not close an app:
* **Solution:** The app might be missing its `exec_path`. Use the **Edit (`e`)** function in the UI to manually add the file path to the executable (e.g., `C:\Program Files\Adobe\Photoshop.exe`).
