package phi.tests;

import static org.junit.Assert.*;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.*;
import java.util.concurrent.TimeUnit;

public final class ConsoleSupport {
    public static String run(String name, String input, String... args) throws Exception {
        String java = Path.of(System.getProperty("java.home"), "bin", "java").toString();
        List<String> command = new ArrayList<>(List.of(java, "-Dfile.encoding=UTF-8", "-Duser.language=en", "-Duser.country=US", "-cp", System.getProperty("java.class.path"), name));
        command.addAll(Arrays.asList(args));
        Process process = new ProcessBuilder(command).redirectErrorStream(true).start();
        ByteArrayOutputStream output = new ByteArrayOutputStream();
        Thread reader = new Thread(() -> {
            try (InputStream stream = process.getInputStream()) {
                byte[] buffer = new byte[4096];
                int count;
                while ((count = stream.read(buffer)) != -1) {
                    if (output.size() < 65536) output.write(buffer, 0, Math.min(count, 65536 - output.size()));
                }
            } catch (IOException ignored) { }
        });
        reader.setDaemon(true); reader.start();
        try {
            try (OutputStream stdin = process.getOutputStream()) { stdin.write(input.getBytes(StandardCharsets.UTF_8)); }
            if (!process.waitFor(4, TimeUnit.SECONDS)) {
                process.destroyForcibly();
                fail(name + " did not finish. Check for an infinite loop or unexpected input prompt.");
            }
            reader.join(1000);
            String text = output.toString(StandardCharsets.UTF_8).replace("\r\n", "\n");
            assertEquals("Run " + name + ": " + text, 0, process.exitValue());
            return text;
        } finally { if (process.isAlive()) process.destroyForcibly(); }
    }
}
