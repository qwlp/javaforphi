package phi.eclipse;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Comparator;
import org.eclipse.core.resources.IProject;
import org.eclipse.core.resources.IProjectDescription;
import org.eclipse.core.resources.IResource;
import org.eclipse.core.resources.IWorkspace;
import org.eclipse.core.resources.ResourcesPlugin;
import org.eclipse.core.runtime.IProgressMonitor;
import org.eclipse.core.runtime.Platform;

public final class Importer {
    static Path queue() throws Exception {
        return Path.of(Platform.getConfigurationLocation().getURL().toURI()).resolve("phi-imports");
    }

    static void process(IProgressMonitor monitor) throws Exception {
        Path queue = queue();
        Files.createDirectories(queue);
        Exception firstFailure = null;
        try (var requests = Files.list(queue)) {
            for (Path request : requests.filter(p -> p.toString().endsWith(".request")).sorted(Comparator.naturalOrder()).toList()) {
                try {
                    Path directory = Path.of(Files.readString(request, StandardCharsets.UTF_8)).toRealPath();
                    IWorkspace workspace = ResourcesPlugin.getWorkspace();
                    String[] importedLocation = new String[1];
                    workspace.run(progress -> {
                        var location = org.eclipse.core.runtime.Path.fromOSString(directory.toString());
                        IProject project = null;
                        for (IProject existing : workspace.getRoot().getProjects()) {
                            if (location.equals(existing.getLocation())) {
                                project = existing;
                                break;
                            }
                        }
                        if (project == null) {
                            IProjectDescription description = workspace.loadProjectDescription(location.append(".project"));
                            String name = description.getName();
                            project = workspace.getRoot().getProject(name);
                            if (project.exists()) {
                                name = directory.getFileName() + "-phi-" + Integer.toUnsignedString(directory.toString().hashCode(), 16);
                                project = workspace.getRoot().getProject(name);
                                if (project.exists() && !location.equals(project.getLocation())) {
                                    throw new org.eclipse.core.runtime.CoreException(new org.eclipse.core.runtime.Status(
                                        org.eclipse.core.runtime.IStatus.ERROR, "phi.eclipse", "Project name collision: " + name));
                                }
                            }
                            if (!project.exists()) {
                                description.setName(name);
                                description.setLocation(location);
                                project.create(description, progress);
                            }
                        }
                        if (!project.isOpen()) project.open(progress);
                        project.refreshLocal(IResource.DEPTH_INFINITE, progress);
                        importedLocation[0] = project.getLocation().toOSString();
                    }, monitor);
                    Files.writeString(Path.of(request + ".done"), importedLocation[0], StandardCharsets.UTF_8);
                    Files.deleteIfExists(Path.of(request + ".error"));
                    Files.delete(request);
                } catch (Exception failure) {
                    Files.writeString(Path.of(request + ".error"), failure.toString(), StandardCharsets.UTF_8);
                    Files.deleteIfExists(request);
                    if (firstFailure == null) firstFailure = failure;
                }
            }
        }
        if (firstFailure != null) throw firstFailure;
    }
}
