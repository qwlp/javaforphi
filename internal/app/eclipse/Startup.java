package phi.eclipse;

import java.nio.file.Files;
import org.eclipse.core.runtime.IProgressMonitor;
import org.eclipse.core.runtime.IStatus;
import org.eclipse.core.runtime.Status;
import org.eclipse.core.runtime.jobs.Job;
import org.eclipse.ui.IStartup;
import org.eclipse.ui.statushandlers.StatusManager;

public final class Startup implements IStartup {
    @Override public void earlyStartup() {
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            try { Files.deleteIfExists(Importer.queue().resolve("heartbeat")); } catch (Exception ignored) { }
        }));
        Job job = new Job("Import Phi labs") {
            @Override protected IStatus run(IProgressMonitor monitor) {
                try {
                    Files.writeString(Importer.queue().resolve("heartbeat"), Long.toString(ProcessHandle.current().pid()));
                    Importer.process(monitor);
                    return Status.OK_STATUS;
                } catch (Exception failure) {
                    Status status = new Status(IStatus.ERROR, "phi.eclipse", "Could not import Phi lab", failure);
                    StatusManager.getManager().handle(status, StatusManager.SHOW | StatusManager.LOG);
                    return status;
                } finally {
                    if (!monitor.isCanceled()) schedule(2000);
                }
            }
        };
        job.setSystem(true);
        job.schedule();
    }
}
