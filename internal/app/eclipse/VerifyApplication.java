package phi.eclipse;

import org.eclipse.core.resources.ResourcesPlugin;
import org.eclipse.equinox.app.IApplication;
import org.eclipse.equinox.app.IApplicationContext;

/** Verifies the plugin can resolve and access the workspace without importing labs. */
public final class VerifyApplication implements IApplication {
    @Override public Object start(IApplicationContext context) throws Exception {
        ResourcesPlugin.getWorkspace().getRoot();
        System.out.println("PHI_ECLIPSE_PLUGIN_READY");
        return IApplication.EXIT_OK;
    }
    @Override public void stop() { }
}
