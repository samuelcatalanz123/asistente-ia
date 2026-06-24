class Message {
  final String role; // "user" or "assistant"
  final String content;
  final String? imageUrl; // si la respuesta trae una imagen generada

  const Message({required this.role, required this.content, this.imageUrl});

  bool get isUser => role == 'user';

  Map<String, dynamic> toJson() => {
        'role': role,
        'content': content,
        if (imageUrl != null) 'imageUrl': imageUrl,
      };

  factory Message.fromJson(Map<String, dynamic> json) => Message(
        role: json['role'] as String,
        content: json['content'] as String,
        imageUrl: json['imageUrl'] as String?,
      );
}
