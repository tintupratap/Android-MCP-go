package com.android.mcp;

import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.Base64;

public class HelperMain {
    public static void main(String[] args) {
        if (args.length < 1) {
            System.err.println("Usage: HelperMain <command> [args...]");
            System.exit(1);
        }
        String cmd = args[0];
        try {
            if ("drag".equalsIgnoreCase(cmd)) {
                doDrag(args);
            } else if ("tap".equalsIgnoreCase(cmd)) {
                doTap(args);
            } else if ("type".equalsIgnoreCase(cmd)) {
                doType(args);
            } else if ("pinch".equalsIgnoreCase(cmd)) {
                doPinch(args);
            } else if ("dump".equalsIgnoreCase(cmd)) {
                doDump(args);
            } else {
                System.err.println("Unknown command: " + cmd);
                System.exit(1);
            }
        } catch (Exception e) {
            e.printStackTrace(System.err);
            System.exit(1);
        }
    }

    private static Object getInputManager() throws Exception {
        Class<?> inputManagerClass = Class.forName("android.hardware.input.InputManager");
        try {
            Method getInstanceMethod = inputManagerClass.getMethod("getInstance");
            return getInstanceMethod.invoke(null);
        } catch (Exception e) {
            Class<?> serviceManagerClass = Class.forName("android.os.ServiceManager");
            Method getServiceMethod = serviceManagerClass.getMethod("getService", String.class);
            Object b = getServiceMethod.invoke(null, "input");
            Class<?> stubClass = Class.forName("android.hardware.input.IInputManager$Stub");
            Method asInterfaceMethod = stubClass.getMethod("asInterface", Class.forName("android.os.IBinder"));
            return asInterfaceMethod.invoke(null, b);
        }
    }

    private static Method getInjectMethod(Object im) {
        for (Method m : im.getClass().getMethods()) {
            if (m.getName().equals("injectInputEvent")) {
                return m;
            }
        }
        return null;
    }

    private static void prepareLooper() {
        try {
            Class<?> looperClass = Class.forName("android.os.Looper");
            Method prepareMethod = looperClass.getMethod("prepareMainLooper");
            prepareMethod.invoke(null);
        } catch (Exception ignored) {}
    }

    private static void doDrag(String[] args) throws Exception {
        if (args.length < 5) {
            System.err.println("Usage: drag <x1> <y1> <x2> <y2> [duration_ms]");
            System.exit(1);
        }
        float x1 = Float.parseFloat(args[1]);
        float y1 = Float.parseFloat(args[2]);
        float x2 = Float.parseFloat(args[3]);
        float y2 = Float.parseFloat(args[4]);
        int duration = args.length >= 6 ? Integer.parseInt(args[5]) : 1000;

        prepareLooper();
        Object im = getInputManager();
        Method injectMethod = getInjectMethod(im);

        Class<?> motionEventClass = Class.forName("android.view.MotionEvent");
        Method obtainMethod = motionEventClass.getMethod("obtain", long.class, long.class, int.class, float.class, float.class, int.class);
        Method recycleMethod = motionEventClass.getMethod("recycle");

        Class<?> systemClockClass = Class.forName("android.os.SystemClock");
        Method uptimeMillisMethod = systemClockClass.getMethod("uptimeMillis");

        long downTime = (Long) uptimeMillisMethod.invoke(null);
        long eventTime = downTime;

        // 1. ACTION_DOWN
        Object event = obtainMethod.invoke(null, downTime, eventTime, 0, x1, y1, 0);
        injectMethod.invoke(im, event, 2);
        recycleMethod.invoke(event);

        // 2. Stationary hold (800ms) for LONG_PRESS
        Thread.sleep(800);

        // 3. ACTION_MOVE interpolation steps
        int steps = 25;
        long stepDelay = Math.max(10, duration / steps);
        for (int i = 1; i <= steps; i++) {
            eventTime = (Long) uptimeMillisMethod.invoke(null);
            float cx = x1 + (x2 - x1) * i / steps;
            float cy = y1 + (y2 - y1) * i / steps;
            event = obtainMethod.invoke(null, downTime, eventTime, 2, cx, cy, 0);
            injectMethod.invoke(im, event, 2);
            recycleMethod.invoke(event);
            Thread.sleep(stepDelay);
        }

        Thread.sleep(200);

        // 4. ACTION_UP
        eventTime = (Long) uptimeMillisMethod.invoke(null);
        event = obtainMethod.invoke(null, downTime, eventTime, 1, x2, y2, 0);
        injectMethod.invoke(im, event, 2);
        recycleMethod.invoke(event);

        System.out.println("MCP_DRAG_SUCCESS");
    }

    private static void doTap(String[] args) throws Exception {
        if (args.length < 3) {
            System.err.println("Usage: tap <x> <y>");
            System.exit(1);
        }
        float x = Float.parseFloat(args[1]);
        float y = Float.parseFloat(args[2]);

        prepareLooper();
        Object im = getInputManager();
        Method injectMethod = getInjectMethod(im);

        Class<?> motionEventClass = Class.forName("android.view.MotionEvent");
        Method obtainMethod = motionEventClass.getMethod("obtain", long.class, long.class, int.class, float.class, float.class, int.class);
        Method recycleMethod = motionEventClass.getMethod("recycle");

        Class<?> systemClockClass = Class.forName("android.os.SystemClock");
        Method uptimeMillisMethod = systemClockClass.getMethod("uptimeMillis");

        long downTime = (Long) uptimeMillisMethod.invoke(null);
        long eventTime = downTime;

        Object event = obtainMethod.invoke(null, downTime, eventTime, 0, x, y, 0);
        injectMethod.invoke(im, event, 2);
        recycleMethod.invoke(event);

        Thread.sleep(50);

        eventTime = (Long) uptimeMillisMethod.invoke(null);
        event = obtainMethod.invoke(null, downTime, eventTime, 1, x, y, 0);
        injectMethod.invoke(im, event, 2);
        recycleMethod.invoke(event);

        System.out.println("MCP_TAP_SUCCESS");
    }

    private static void doType(String[] args) throws Exception {
        if (args.length < 2) {
            System.err.println("Usage: type <base64_text> [x] [y]");
            System.exit(1);
        }
        String text = new String(Base64.getDecoder().decode(args[1]), StandardCharsets.UTF_8);

        if (args.length >= 4) {
            float x = Float.parseFloat(args[2]);
            float y = Float.parseFloat(args[3]);
            if (x > 0 && y > 0) {
                doTap(new String[]{"tap", String.valueOf(x), String.valueOf(y)});
                Thread.sleep(200);
            }
        }

        prepareLooper();
        Object im = getInputManager();
        Method injectMethod = getInjectMethod(im);

        Class<?> keyCharacterMapClass = Class.forName("android.view.KeyCharacterMap");
        Method getEventsMethod = keyCharacterMapClass.getMethod("getEvents", char[].class);

        char[] chars = text.toCharArray();
        Object eventsObj = getEventsMethod.invoke(null, (Object) chars);

        if (eventsObj != null) {
            Object[] events = (Object[]) eventsObj;
            for (Object ke : events) {
                injectMethod.invoke(im, ke, 2);
            }
        } else {
            // Fallback
            Process p = Runtime.getRuntime().exec(new String[]{"input", "text", text});
            p.waitFor();
        }

        System.out.println("MCP_TYPE_SUCCESS");
    }

    private static void doPinch(String[] args) throws Exception {
        if (args.length < 9) {
            System.err.println("Usage: pinch <x1> <y1> <x2> <y2> <x3> <y3> <x4> <y4> [duration_ms]");
            System.exit(1);
        }
        float x1 = Float.parseFloat(args[1]);
        float y1 = Float.parseFloat(args[2]);
        float x2 = Float.parseFloat(args[3]);
        float y2 = Float.parseFloat(args[4]);
        float x3 = Float.parseFloat(args[5]);
        float y3 = Float.parseFloat(args[6]);
        float x4 = Float.parseFloat(args[7]);
        float y4 = Float.parseFloat(args[8]);
        int duration = args.length >= 10 ? Integer.parseInt(args[9]) : 500;

        prepareLooper();
        Object im = getInputManager();
        Method injectMethod = getInjectMethod(im);

        Class<?> motionEventClass = Class.forName("android.view.MotionEvent");
        Class<?> pointerPropertiesClass = Class.forName("android.view.MotionEvent$PointerProperties");
        Class<?> pointerCoordsClass = Class.forName("android.view.MotionEvent$PointerCoords");

        Object pp0 = pointerPropertiesClass.newInstance();
        pointerPropertiesClass.getField("id").setInt(pp0, 0);
        Object pp1 = pointerPropertiesClass.newInstance();
        pointerPropertiesClass.getField("id").setInt(pp1, 1);

        Object ppArray = java.lang.reflect.Array.newInstance(pointerPropertiesClass, 2);
        java.lang.reflect.Array.set(ppArray, 0, pp0);
        java.lang.reflect.Array.set(ppArray, 1, pp1);

        Object pc0 = pointerCoordsClass.newInstance();
        pointerCoordsClass.getField("x").setFloat(pc0, x1);
        pointerCoordsClass.getField("y").setFloat(pc0, y1);
        pointerCoordsClass.getField("pressure").setFloat(pc0, 1.0f);
        pointerCoordsClass.getField("size").setFloat(pc0, 1.0f);

        Object pc1 = pointerCoordsClass.newInstance();
        pointerCoordsClass.getField("x").setFloat(pc1, x3);
        pointerCoordsClass.getField("y").setFloat(pc1, y3);
        pointerCoordsClass.getField("pressure").setFloat(pc1, 1.0f);
        pointerCoordsClass.getField("size").setFloat(pc1, 1.0f);

        Object pcArray = java.lang.reflect.Array.newInstance(pointerCoordsClass, 2);
        java.lang.reflect.Array.set(pcArray, 0, pc0);
        java.lang.reflect.Array.set(pcArray, 1, pc1);

        Method obtainMultiMethod = null;
        for (Method m : motionEventClass.getMethods()) {
            if (m.getName().equals("obtain")) {
                Class<?>[] types = m.getParameterTypes();
                if (types.length >= 13 && types[4].isArray() && types[4].getComponentType().getName().contains("PointerProperties")) {
                    obtainMultiMethod = m;
                    break;
                }
            }
        }

        Class<?> systemClockClass = Class.forName("android.os.SystemClock");
        Method uptimeMillisMethod = systemClockClass.getMethod("uptimeMillis");

        long downTime = (Long) uptimeMillisMethod.invoke(null);
        long eventTime = downTime;

        // Pointer 0 ACTION_DOWN (0)
        Object singlePP = java.lang.reflect.Array.newInstance(pointerPropertiesClass, 1);
        java.lang.reflect.Array.set(singlePP, 0, pp0);
        Object singlePC = java.lang.reflect.Array.newInstance(pointerCoordsClass, 1);
        java.lang.reflect.Array.set(singlePC, 0, pc0);

        Object event = obtainMultiMethod.invoke(null, downTime, eventTime, 0, 1, singlePP, singlePC, 0, 0, 1.0f, 1.0f, 0, 0, 4098, 0);
        injectMethod.invoke(im, event, 2);

        // Pointer 1 ACTION_POINTER_DOWN (0x0105 = 261)
        eventTime = (Long) uptimeMillisMethod.invoke(null);
        event = obtainMultiMethod.invoke(null, downTime, eventTime, 261, 2, ppArray, pcArray, 0, 0, 1.0f, 1.0f, 0, 0, 4098, 0);
        injectMethod.invoke(im, event, 2);

        // Interpolate steps
        int steps = 20;
        long stepDelay = Math.max(10, duration / steps);
        for (int i = 1; i <= steps; i++) {
            eventTime = (Long) uptimeMillisMethod.invoke(null);
            float cx0 = x1 + (x2 - x1) * i / steps;
            float cy0 = y1 + (y2 - y1) * i / steps;
            float cx1 = x3 + (x4 - x3) * i / steps;
            float cy1 = y3 + (y4 - y3) * i / steps;

            pointerCoordsClass.getField("x").setFloat(pc0, cx0);
            pointerCoordsClass.getField("y").setFloat(pc0, cy0);
            pointerCoordsClass.getField("x").setFloat(pc1, cx1);
            pointerCoordsClass.getField("y").setFloat(pc1, cy1);

            event = obtainMultiMethod.invoke(null, downTime, eventTime, 2, 2, ppArray, pcArray, 0, 0, 1.0f, 1.0f, 0, 0, 4098, 0);
            injectMethod.invoke(im, event, 2);
            Thread.sleep(stepDelay);
        }

        // Pointer 1 ACTION_POINTER_UP (0x0106 = 262)
        eventTime = (Long) uptimeMillisMethod.invoke(null);
        event = obtainMultiMethod.invoke(null, downTime, eventTime, 262, 2, ppArray, pcArray, 0, 0, 1.0f, 1.0f, 0, 0, 4098, 0);
        injectMethod.invoke(im, event, 2);

        // Pointer 0 ACTION_UP (1)
        eventTime = (Long) uptimeMillisMethod.invoke(null);
        event = obtainMultiMethod.invoke(null, downTime, eventTime, 1, 1, singlePP, singlePC, 0, 0, 1.0f, 1.0f, 0, 0, 4098, 0);
        injectMethod.invoke(im, event, 2);

        System.out.println("MCP_PINCH_SUCCESS");
    }

    private static void doDump(String[] args) throws Exception {
        Process p = Runtime.getRuntime().exec(new String[]{"uiautomator", "dump", "/data/local/tmp/ui_dump.xml"});
        p.waitFor();
        Process catProc = Runtime.getRuntime().exec(new String[]{"cat", "/data/local/tmp/ui_dump.xml"});
        java.io.InputStream is = catProc.getInputStream();
        byte[] buf = new byte[8192];
        int len;
        while ((len = is.read(buf)) != -1) {
            System.out.write(buf, 0, len);
        }
        System.out.flush();
    }
}
