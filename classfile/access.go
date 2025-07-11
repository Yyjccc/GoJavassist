package classfile

//访问标志符

// C=class
// c=inner_class_access_flags
// F=field
// M=method
// P=MethodParameter
// D=module
// R=requires_flags
// X=exports_flags
// O=opens_flags
const (
	AccPublic       = 0x0001 //      CcFM_____			public
	AccPrivate      = 0x0002 //      _cFM_____			private
	AccProtected    = 0x0004 //      _cFM_____			protected
	AccStatic       = 0x0008 //      _cFM_____			static
	AccFinal        = 0x0010 //      CcFMP____			final
	AccSuper        = 0x0020 //      C________			super
	AccSynchronized = 0x0020 //      ___M_____			synchronized
	AccOpen         = 0x0020 // 9,   _____D___			open
	AccTransitive   = 0x0020 // 9,   ______R__			transitive
	AccVolatile     = 0x0040 //      __F______			volatile
	AccBridge       = 0x0040 //      ___M_____			bridge
	AccStaticPhase  = 0x0040 // 9,   ______R__			staticPhase
	AccTransient    = 0x0080 //      __F______		transient
	AccVarargs      = 0x0080 // 5.0  ___M_____		varargs
	AccNative       = 0x0100 //      ___M_____		native
	AccInterface    = 0x0200 //      Cc_______		interface
	AccAbstract     = 0x0400 //      Cc_M_____		abstract
	AccStrict       = 0x0800 //      ___M_____		strict
	AccSynthetic    = 0x1000 //      CcFMPDRXO		synthetic
	AccAnnotation   = 0x2000 // 5.0, Cc_______		annotation
	AccEnum         = 0x4000 // 5.0, CcF______		enum
	AccModule       = 0x8000 // 9,   C________		module
	AccMandated     = 0x8000 // ?,   ____PDRXO		mandated
)

type AccessFlags uint16

func (flags AccessFlags) IsPublic() bool       { return flags&AccPublic != 0 }
func (flags AccessFlags) IsPrivate() bool      { return flags&AccPrivate != 0 }
func (flags AccessFlags) IsProtected() bool    { return flags&AccProtected != 0 }
func (flags AccessFlags) IsStatic() bool       { return flags&AccStatic != 0 }
func (flags AccessFlags) IsFinal() bool        { return flags&AccFinal != 0 }
func (flags AccessFlags) IsSuper() bool        { return flags&AccSuper != 0 }
func (flags AccessFlags) IsSynchronized() bool { return flags&AccSynchronized != 0 }
func (flags AccessFlags) IsOpen() bool         { return flags&AccOpen != 0 }
func (flags AccessFlags) IsTransitive() bool   { return flags&AccTransitive != 0 }
func (flags AccessFlags) IsVolatile() bool     { return flags&AccVolatile != 0 }
func (flags AccessFlags) IsBridge() bool       { return flags&AccBridge != 0 }
func (flags AccessFlags) IsStaticPhase() bool  { return flags&AccStaticPhase != 0 }
func (flags AccessFlags) IsTransient() bool    { return flags&AccTransient != 0 }
func (flags AccessFlags) IsVarargs() bool      { return flags&AccVarargs != 0 }
func (flags AccessFlags) IsNative() bool       { return flags&AccNative != 0 }
func (flags AccessFlags) IsInterface() bool    { return flags&AccInterface != 0 }
func (flags AccessFlags) IsAbstract() bool     { return flags&AccAbstract != 0 }
func (flags AccessFlags) IsStrict() bool       { return flags&AccStrict != 0 }
func (flags AccessFlags) IsSynthetic() bool    { return flags&AccSynthetic != 0 }
func (flags AccessFlags) IsAnnotation() bool   { return flags&AccAnnotation != 0 }
func (flags AccessFlags) IsEnum() bool         { return flags&AccEnum != 0 }
func (flags AccessFlags) IsModule() bool       { return flags&AccModule != 0 }
func (flags AccessFlags) IsMandated() bool     { return flags&AccMandated != 0 }
