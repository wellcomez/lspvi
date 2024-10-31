function aaa(){
    console.log("")
}
// 定义类
class Person {
    // 构造函数
    constructor(name, age) {
        this.name = name;
        this.age = age;
    }

    // 方法
    introduce() {
        console.log(`Hello, my name is ${this.name} and I am ${this.age} years old.`);
    }
}

// 创建对象
const person1 = new Person('Alice', 30);
person1.introduce(); // 输出: Hello, my name is Alice and I am 30 years old.