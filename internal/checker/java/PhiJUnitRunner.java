package phi.runner;

import java.io.PrintStream;
import java.util.HashSet;
import java.util.Set;
import org.junit.runner.Description;
import org.junit.runner.JUnitCore;
import org.junit.runner.Result;
import org.junit.runner.notification.Failure;
import org.junit.runner.notification.RunListener;

/** Instructor-owned runner. Case events are separate from learner output. */
public final class PhiJUnitRunner {
    public static void main(String[] args) throws Exception {
        final PrintStream events = System.out;
        final Set<String> failed = new HashSet<>();
        JUnitCore junit = new JUnitCore();
        junit.addListener(new RunListener() {
            private void event(String state, Description test, String detail) {
                events.println("PHI_CASE\t" + state + "\t" + clean(test.getDisplayName()) + "\t" + clean(detail));
                events.flush();
            }
            public void testStarted(Description test) { event("running", test, ""); }
            public void testFailure(Failure failure) {
                failed.add(failure.getDescription().getDisplayName());
                event("failed", failure.getDescription(), failure.getMessage());
            }
            public void testAssumptionFailure(Failure failure) {
                failed.add(failure.getDescription().getDisplayName());
                event("failed", failure.getDescription(), "Required grading case was skipped: " + failure.getMessage());
            }
            public void testIgnored(Description test) {
                failed.add(test.getDisplayName());
                event("failed", test, "Required grading case was ignored");
            }
            public void testFinished(Description test) {
                if (!failed.contains(test.getDisplayName())) event("passed", test, "");
            }
        });
        Class<?>[] suites = new Class<?>[args.length];
        for (int i = 0; i < args.length; i++) suites[i] = Class.forName(args[i]);
        Result result = junit.run(suites);
        for (Failure failure : result.getFailures()) {
            events.println("\n" + failure.getDescription().getDisplayName());
            events.println(failure.getTrace());
        }
        events.println("PHI_TOTAL\t" + result.getRunCount() + "\t" + failed.size());
        // A suite with no executed cases cannot establish completion.
        System.exit(result.wasSuccessful() && failed.isEmpty() && result.getRunCount() > 0 ? 0 : 1);
    }
    private static String clean(String value) {
        return value == null ? "" : value.replace('\t', ' ').replace('\n', ' ').replace('\r', ' ');
    }
}
