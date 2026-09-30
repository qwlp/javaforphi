package phi.eclipse;

import org.eclipse.core.runtime.NullProgressMonitor;
import org.eclipse.core.resources.ResourcesPlugin;
import org.eclipse.equinox.app.IApplication;
import org.eclipse.equinox.app.IApplicationContext;

// Used by the integration test to exercise the same importer without a GUI.
public final class ImportApplication implements IApplication {
    @Override public Object start(IApplicationContext context) throws Exception {
        Importer.process(new NullProgressMonitor());
        ResourcesPlugin.getWorkspace().save(true, new NullProgressMonitor());
        return IApplication.EXIT_OK;
    }
    @Override public void stop() { }
}
